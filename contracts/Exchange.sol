pragma solidity ^0.8.20;

import "@openzeppelin/contracts/access/Ownable.sol";
import { ECDSA } from "@openzeppelin/contracts/utils/cryptography/ECDSA.sol";
import { IERC20 } from "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import "hardhat/console.sol";
import "./ITokenRegistry.sol";

contract Exchange is Ownable {
    using ECDSA for bytes32;
    
    ITokenRegistry public symbolRegistry;

    // Structs
    // Order represents an order that is off-chain but verified on-chain
    struct Order {
        address createdBy;
        string symbolIn;   
        string symbolOut;  
        uint256 amtIn;     
        uint256 amtOut;    
        uint256 nonce;
    }

    struct TokenPair {
        address makerTokenOut;  
        address takerTokenOut;  
    }
    
    struct SwapInfo {
        Order order;
        bytes signature;
    }

    // Map createdBy's address with an order's nonce (unique order) to represent the filled amount (for order partial fill)
    mapping(address => mapping(uint256 => uint256)) public filledOrdersAmtIn;

    // EIP-712 Type Hash for the Order struct
    bytes32 public constant ORDER_TYPEHASH = keccak256(
        "Order(address createdBy,string symbolIn,string symbolOut,uint256 amtIn,uint256 amtOut,uint256 nonce)"
    );

    bytes32 public DOMAIN_SEPARATOR;

    constructor(address registryAddr) Ownable(msg.sender) {
        symbolRegistry = ITokenRegistry(registryAddr);
        // EIP-712 Domain Separator calculation
        DOMAIN_SEPARATOR = keccak256(
        abi.encode(
            keccak256(
                "EIP712Domain(string name,string version,uint256 chainId,address verifyingContract)"
            ),
            keccak256(bytes("SmashDEX")),
            keccak256(bytes("1")),
            block.chainid,
            address(this)
            )
        );
    }

    // Events
    event OrderExecuted(
        address indexed maker,
        address indexed taker,
        string symbolIn,
        string symbolOut,
        uint256 amtIn,
        uint256 amtOut,
        uint256 makerFilledAmtIn, // Amount of maker's tokenIn actually transferred
        uint256 takerFilledAmtIn  // Amount of taker's tokenIn actually transferred
    );
    event RingTradeExecuted(
        address[] makers,
        string[] symbolsIn,
        string[] symbolsOut,
        uint256[] fillAmounts,
        uint256 ringSize
    );

    // Functions
    // Computes hash for order for signing and verification using EIP-712
    function computeOrderHash(Order memory order) public view returns (bytes32) {
        bytes32 structHash = keccak256(
            abi.encode(
                ORDER_TYPEHASH,
                order.createdBy,
                keccak256(bytes(order.symbolIn)),
                keccak256(bytes(order.symbolOut)),
                order.amtIn,
                order.amtOut,
                order.nonce
            )
        );

        return keccak256(
            abi.encodePacked("\x19\x01", DOMAIN_SEPARATOR, structHash)
        );
    }

    function executeOrder(
        SwapInfo calldata makerInfo, 
        SwapInfo calldata takerInfo,
        uint256 fillAmtIn
    ) external {
        TokenPair memory tokens = TokenPair({
            makerTokenOut: symbolRegistry.getTokenAddress(makerInfo.order.symbolOut),
            takerTokenOut: symbolRegistry.getTokenAddress(takerInfo.order.symbolOut)
        });

        _validityChecks(makerInfo.order, makerInfo.signature, takerInfo.order, takerInfo.signature, fillAmtIn);

        uint256 makerAmtOut = (fillAmtIn * makerInfo.order.amtOut) / makerInfo.order.amtIn;
        
        filledOrdersAmtIn[makerInfo.order.createdBy][makerInfo.order.nonce] += fillAmtIn;
        filledOrdersAmtIn[takerInfo.order.createdBy][takerInfo.order.nonce] += makerAmtOut;

        // Perform token transfer
        _swapTokens(
            makerInfo.order.createdBy,
            tokens.makerTokenOut,
            takerInfo.order.createdBy,
            tokens.takerTokenOut, 
            fillAmtIn,           
            makerAmtOut          
        );

        // Emit event
        emit OrderExecuted(
            makerInfo.order.createdBy,
            takerInfo.order.createdBy,
            makerInfo.order.symbolIn,
            makerInfo.order.symbolOut,
            makerInfo.order.amtIn,
            makerInfo.order.amtOut,
            fillAmtIn,
            makerAmtOut
        );
    }

    function _validityChecks(
        Order calldata maker, 
        bytes calldata makerSignature, 
        Order calldata taker, 
        bytes calldata takerSignature,
        uint256 fillAmtIn
    ) private view {
        // Signature checks
        _validEOACheck(maker, makerSignature);
        _validEOACheck(taker, takerSignature);

        // Token pair and price checks
        _tokenMatch(maker, taker);
        _partialFillMatch(maker, taker);
        
        // Remaining amount and fill size checks
        _checkRemainingAmounts(maker, taker, fillAmtIn);
    }

    function _validEOACheck(Order calldata order, bytes calldata signature) private view {
        bytes32 digest = computeOrderHash(order);
        address recovered = ECDSA.recover(digest, signature);
        require(recovered == order.createdBy, "Invalid order signature");
    }

    function _swapTokens(
        address makerAddress,
        address makerTokenOut,
        address takerAddress,
        address takerTokenOut,
        uint256 makerFillAmt,
        uint256 takerFillAmt
    ) private {
        // Maker sends their tokenOut to Taker (taker receives takerFillAmt)
        require(
            IERC20(makerTokenOut).transferFrom(makerAddress, takerAddress, takerFillAmt),
            "Maker Transfer fail"
        );
        // Taker sends their tokenOut to Maker (maker receives makerFillAmt)
        require(
            IERC20(takerTokenOut).transferFrom(takerAddress, makerAddress, makerFillAmt),
            "Taker Transfer fail"
        );
    }

    // Ensures the maker and taker orders are for the same tokens in opposite directions
    function _tokenMatch(Order calldata maker, Order calldata taker) private pure {
        require(
            keccak256(bytes(maker.symbolIn)) == keccak256(bytes(taker.symbolOut)) &&
            keccak256(bytes(maker.symbolOut)) == keccak256(bytes(taker.symbolIn)),
            "Token mismatch"
        );
    }

    // Ensures the orders have the same price ratio
    function _partialFillMatch(Order calldata maker, Order calldata taker) private pure {
        // Price check maker.amtIn/maker.amtOut should equal taker.amtOut/taker.amtIn
        // Opposite sides should match
        require(
            maker.amtIn * taker.amtIn == maker.amtOut * taker.amtOut,
            "Price mismatch"
        );
    }

    // Checks if the requested fill amount is valid against the remaining unfilled amounts
    function _checkRemainingAmounts(Order calldata maker, Order calldata taker, uint256 fillAmtIn) private view {
        uint256 makerRemaining = getRemainingAmt(maker.createdBy, maker.nonce, maker.amtIn);
        uint256 makerAmtOut = (fillAmtIn * maker.amtOut) / maker.amtIn;
        uint256 takerRemaining = getRemainingAmt(taker.createdBy, taker.nonce, taker.amtIn);

        require(fillAmtIn > 0, "Fill amount must be greater than 0");
        require(fillAmtIn <= makerRemaining, "Fill amount exceeds maker's remaining order");
        require(makerAmtOut <= takerRemaining, "Fill amount exceeds taker's remaining order");
    }

    // Returns remaining amount to fill for a given order (originalAmtIn is required)
    function getRemainingAmt(address createdBy, uint256 nonce, uint256 originalAmtIn) public view returns (uint256) {
        uint256 filled = filledOrdersAmtIn[createdBy][nonce];
        require(filled <= originalAmtIn, "Filled amount exceeds original"); 
        return originalAmtIn - filled;
    }

    function executeRingTrade(
    Order[] calldata ringOrders,
    uint256[] calldata fillAmounts
) external onlyOwner {
    require(ringOrders.length == fillAmounts.length, "Length mismatch");
    
    // Execute all transfers in the ring
    for (uint256 i = 0; i < ringOrders.length; i++) {
        uint256 nextIndex = (i + 1) % ringOrders.length;
        
        Order calldata currentOrder = ringOrders[i];
        Order calldata nextOrder = ringOrders[nextIndex];
        
        // Current order gives their tokenOut to next order's creator
        address tokenOut = symbolRegistry.getTokenAddress(currentOrder.symbolOut);
        uint256 outputAmount = (fillAmounts[i] * currentOrder.amtOut) / currentOrder.amtIn;
        
        // Transfer from current order's creator to next order's creator
        require(
            IERC20(tokenOut).transferFrom(
                currentOrder.createdBy,
                nextOrder.createdBy,
                outputAmount
            ),
            "Ring transfer failed"
        );
        
        // Update filled amounts
        filledOrdersAmtIn[currentOrder.createdBy][currentOrder.nonce] += fillAmounts[i];
    }
    
    address[] memory makers = new address[](ringOrders.length);
    string[] memory symbolsIn = new string[](ringOrders.length);
    string[] memory symbolsOut = new string[](ringOrders.length);
    
    for (uint256 i = 0; i < ringOrders.length; i++) {
        makers[i] = ringOrders[i].createdBy;
        symbolsIn[i] = ringOrders[i].symbolIn;
        symbolsOut[i] = ringOrders[i].symbolOut;
    }
    
    emit RingTradeExecuted(makers, symbolsIn, symbolsOut, fillAmounts, ringOrders.length);
}
}