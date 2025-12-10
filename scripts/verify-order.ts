import { network } from "hardhat";
const { viem } = await network.connect();
const [deployer, scDeployer, dcDeployer, maker, taker, taker2] = await viem.getWalletClients();
const toMint = 1_000_000n * 10n ** 18n;

// Deploy three tokens for ring matching
const smashCoin = await viem.deployContract("SmashCoin", [toMint]);
const doogeCoin = await viem.deployContract("DoogeCoin", [toMint]);
const wEth = await viem.deployContract("WEthereum", [toMint]);

console.log("===EOA===")
console.log(`Deployer: ${deployer.account.address}`)
console.log(`Maker: ${maker.account.address}`)
console.log(`Taker: ${taker.account.address}`)
console.log(`Taker2: ${taker2.account.address}`)
console.log("=========")

const jsonString = (payload: any) => {
    return JSON.stringify(payload, (_, value) => {
        if (typeof value === 'bigint') {
            return value.toString();
        }
        return value;
    });
}

// Get nonces for all three users
const nonceResp1 = await fetch('http://localhost:11223/nonce', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: jsonString({address: maker.account.address})
})
const nonceMaker = (await nonceResp1.json()).nonce;

const nonceResp2 = await fetch('http://localhost:11223/nonce', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: jsonString({address: taker.account.address})
})
const nonceTaker = (await nonceResp2.json()).nonce;

const nonceResp3 = await fetch('http://localhost:11223/nonce', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: jsonString({address: taker2.account.address})
})
const nonceTaker2 = (await nonceResp3.json()).nonce;

// Deploy registry
const registry = await viem.deployContract("TokenRegistry");
await registry.write.addToken([smashCoin.address]);
await registry.write.addToken([doogeCoin.address]);
await registry.write.addToken([wEth.address]);

// Distribute tokens according to your setup:
// Maker gets SmashCoin (SC)
// Taker gets DoogeCoin (DC)
// Taker2 gets WETH
const toSend = 1_000n * 10n ** 18n;
await smashCoin.write.transfer([maker.account.address, toSend]);
await doogeCoin.write.transfer([taker.account.address, toSend]);
await wEth.write.transfer([taker2.account.address, toSend]);

const dex = await viem.deployContract("Exchange", [registry.address]);

const chainId = await deployer.getChainId();
console.log(`DEX Address: ${dex.address}`);
console.log(`SmashCoin Address: ${smashCoin.address}`);
console.log(`DoogeCoin Address: ${doogeCoin.address}`);
console.log(`WETH Address: ${wEth.address}`);

// ---------------------------
// Ring Orders: SC -> DC -> WETH -> SC
// Maker: wants to trade SC for DC
// Taker: wants to trade DC for WETH
// Taker2: wants to trade WETH for SC (completes the ring!)
// ---------------------------

const makerOrder = {
    createdBy: maker.account.address,
    symbolIn: "SC",
    symbolOut: "DC",
    amtIn: 300n * 10n ** 18n,      // Maker gives 300 SC
    amtOut: 200n * 10n ** 18n,     // Maker wants 200 DC (rate: 1.5 SC per DC)
    nonce: BigInt(nonceMaker),
};

const takerOrder = {
    createdBy: taker.account.address,
    symbolIn: "DC",
    symbolOut: "WETH",
    amtIn: 200n * 10n ** 18n,      // Taker gives 200 DC
    amtOut: 100n * 10n ** 18n,     // Taker wants 100 WETH (rate: 2 DC per WETH)
    nonce: BigInt(nonceTaker),
};

const taker2Order = {
    createdBy: taker2.account.address,
    symbolIn: "WETH",
    symbolOut: "SC",
    amtIn: 100n * 10n ** 18n,      // Taker2 gives 100 WETH
    amtOut: 300n * 10n ** 18n,     // Taker2 wants 300 SC (rate: 3 SC per WETH)
    nonce: BigInt(nonceTaker2),
};

// ---------------------------
// Helper: EIP-712 signing for orders
// ---------------------------
async function signOrder(order: any, signer: any) {
    const domain = {
        name: "SmashDEX",
        version: "1",
        chainId,
        verifyingContract: "0x9fE46736679d2D9a65F0992F2272dE9f3c7fa6e0",
    };

    const types = {
        Order: [
            { name: "createdBy", type: "address" },
            { name: "symbolIn", type: "string" },
            { name: "symbolOut", type: "string" },
            { name: "amtIn", type: "uint256" },
            { name: "amtOut", type: "uint256" },
            { name: "nonce", type: "uint256" },
        ],
    };

    const signature = await signer.signTypedData({ 
        domain, 
        types, 
        primaryType: "Order", 
        message: order 
    });

    return signature as `0x${string}`;
}

const makerSignature = await signOrder(makerOrder, maker);
const takerSignature = await signOrder(takerOrder, taker);
const taker2Signature = await signOrder(taker2Order, taker2);

const makerPayload = {
    order: makerOrder,
    signature: makerSignature,
}

const takerPayload = {
    order: takerOrder,
    signature: takerSignature,
}

const taker2Payload = {
    order: taker2Order,
    signature: taker2Signature,
}

// Submit all three orders to create a ring
console.log("\n=== SUBMITTING RING ORDERS ===");

console.log("\n--- Order 1: Maker (SC -> DC) ---");
console.log(jsonString(makerPayload));
const response1 = await fetch('http://localhost:11223/order/limit', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: jsonString(makerPayload)
});
console.log(`Status: ${response1.status}`);
await new Promise(resolve => setTimeout(resolve, 500)); // Small delay

console.log("\n--- Order 2: Taker (DC -> WETH) ---");
console.log(jsonString(takerPayload));
const response2 = await fetch('http://localhost:11223/order/limit', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: jsonString(takerPayload)
});
console.log(`Status: ${response2.status}`);
await new Promise(resolve => setTimeout(resolve, 500)); // Small delay

console.log("\n--- Order 3: Taker2 (WETH -> SC) [RING COMPLETES!] ---");
console.log(jsonString(taker2Payload));
const response3 = await fetch('http://localhost:11223/order/limit', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: jsonString(taker2Payload)
});
console.log(`Status: ${response3.status}`);

console.log("\n=== RING MATCHING COMPLETE ===");
console.log("Check your backend logs for ring execution details!");
console.log("\nExpected Ring Path:");
console.log("SC (300) -> DC (200) -> WETH (100) -> SC (300)");
console.log("\nBottleneck calculation:");
console.log("- Maker can provide: 300 SC");
console.log("- Taker needs 200 DC, gets 200 DC from maker (maker uses 300 SC)");
console.log("- Taker2 needs 100 WETH, gets 100 WETH from taker (taker uses 200 DC)");
console.log("- Maker needs 200 DC, gets 200 DC from taker2's output");
console.log("Ring should execute with bottleneck = 200 DC");

// Optional: Query orders for verification
await new Promise(resolve => setTimeout(resolve, 2000));
console.log("\n--- Querying Final Order States ---");

const getMakerOrders = await fetch(`http://localhost:11223/orders/${maker.account.address}`, {
    method: "GET"
});
console.log("\nMaker Orders:", await getMakerOrders.json());

const getTakerOrders = await fetch(`http://localhost:11223/orders/${taker.account.address}`, {
    method: "GET"
});
console.log("\nTaker Orders:", await getTakerOrders.json());

const getTaker2Orders = await fetch(`http://localhost:11223/orders/${taker2.account.address}`, {
    method: "GET"
});
console.log("\nTaker2 Orders:", await getTaker2Orders.json());
