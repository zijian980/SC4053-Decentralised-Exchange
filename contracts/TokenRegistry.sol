// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "@openzeppelin/contracts/access/Ownable.sol";
import "@openzeppelin/contracts/token/ERC20/extensions/IERC20Metadata.sol";

contract TokenRegistry is Ownable {
    struct TokenDetails {
        string name;
        string symbol;
        uint8 decimals;
        bool exists;
    }


    mapping(address => TokenDetails) private tokens;
    mapping(string => address) private symbolToAddress;
    address[] private tokenList;

    address constant ETH_ADDRESS = address(0);

    constructor() Ownable(msg.sender) {
        //tokens[address(0)] = TokenDetails({name: "Ethereum", symbol: "ETH", decimals: 18, exists: true});
        //tokenList.push(address(0));
        //symbolToAddress["ETH"] = ETH_ADDRESS;
    }

    // Events
    event TokenAdded(address indexed token, string name, string symbol);
    event TokenRemoved(address indexed token, string name, string symbol);

    // Functions
    function addToken(address token) external onlyOwner {
        require(token != address(0), "Invalid token address"); // address(0) refers to 0x0
        require(bytes(tokens[token].name).length == 0, "Token already added");
        
        string memory name = IERC20Metadata(token).name();
        string memory symbol = IERC20Metadata(token).symbol();
        require(symbolToAddress[symbol] == address(0), "Symbol already used");

        tokens[token] = TokenDetails({name: name, symbol: symbol, decimals: IERC20Metadata(token).decimals(), exists: true});
        symbolToAddress[symbol] = token;
        tokenList.push(token);

        emit TokenAdded(token, name, symbol);
    }

    function removeToken(address token) external onlyOwner {
        require(bytes(tokens[token].name).length != 0, "Token not supported");

        string memory symbol = tokens[token].symbol;
        
        delete symbolToAddress[symbol];
        delete tokens[token];

        for (uint i = 0; i < tokenList.length; i++) {
            if (tokenList[i] == token) {
                tokenList[i] = tokenList[tokenList.length - 1]; // Swap and pop
                tokenList.pop();
                break;
            }
        }

        emit TokenRemoved(token, tokens[token].name, tokens[token].symbol);
    }

    function isSupported(address token) external view returns (bool) {
        return tokens[token].exists;
    }

    function getAllTokens() external view returns (address[] memory) {
        return tokenList;
    }

    function getTokenAddress(string calldata symbol) external view returns (address) {
        return symbolToAddress[symbol];
    }

}
