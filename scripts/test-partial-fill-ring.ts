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

// Distribute tokens
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


const makerOrder = {
    createdBy: maker.account.address,
    symbolIn: "SC",
    symbolOut: "DC",
    amtIn: 600n * 10n ** 18n,      // Maker gives 600 SC
    amtOut: 400n * 10n ** 18n,     // Maker wants 400 DC (rate: 1.5 SC per DC)
    nonce: BigInt(nonceMaker),
};

const takerOrder = {
    createdBy: taker.account.address,
    symbolIn: "DC",
    symbolOut: "WETH",
    amtIn: 200n * 10n ** 18n,      // Taker gives 200 DC (BOTTLENECK!)
    amtOut: 100n * 10n ** 18n,     // Taker wants 100 WETH (rate: 2 DC per WETH)
    nonce: BigInt(nonceTaker),
};

const taker2Order = {
    createdBy: taker2.account.address,
    symbolIn: "WETH",
    symbolOut: "SC",
    amtIn: 150n * 10n ** 18n,      // Taker2 gives 150 WETH
    amtOut: 450n * 10n ** 18n,     // Taker2 wants 450 SC (rate: 3 SC per WETH)
    nonce: BigInt(nonceTaker2),
};

// Helper: EIP-712 signing for orders
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
console.log("\n=== SUBMITTING PARTIAL FILL RING TEST ===");
console.log("\nExpected Behavior:");
console.log("- Maker: 600 SC -> 400 DC (will fill 300 SC, 50% filled)");
console.log("- Taker: 200 DC -> 100 WETH (BOTTLENECK, will fill 100%, fully filled)");
console.log("- Taker2: 150 WETH -> 450 SC (will fill 100 WETH, 66.67% filled)");
console.log("- Bottleneck: 200 DC (limited by Taker's order size)\n");

console.log("\n--- Order 1: Maker (SC -> DC, LARGE) ---");
console.log(`Amount: ${makerOrder.amtIn.toString()} SC -> ${makerOrder.amtOut.toString()} DC`);
const response1 = await fetch('http://localhost:11223/order/limit', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: jsonString(makerPayload)
});
console.log(`Status: ${response1.status}`);
await new Promise(resolve => setTimeout(resolve, 500));

console.log("\n--- Order 2: Taker (DC -> WETH, BOTTLENECK) ---");
console.log(`Amount: ${takerOrder.amtIn.toString()} DC -> ${takerOrder.amtOut.toString()} WETH`);
const response2 = await fetch('http://localhost:11223/order/limit', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: jsonString(takerPayload)
});
console.log(`Status: ${response2.status}`);
await new Promise(resolve => setTimeout(resolve, 500));

console.log("\n--- Order 3: Taker2 (WETH -> SC, MEDIUM) ---");
console.log(`Amount: ${taker2Order.amtIn.toString()} WETH -> ${taker2Order.amtOut.toString()} SC`);
const response3 = await fetch('http://localhost:11223/order/limit', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: jsonString(taker2Payload)
});
console.log(`Status: ${response3.status}`);

console.log("\n=== WAITING FOR RING EXECUTION ===");
await new Promise(resolve => setTimeout(resolve, 3000));

console.log("\n--- Querying Final Order States ---");

const getMakerOrders = await fetch(`http://localhost:11223/orders/${maker.account.address}`, {
    method: "GET"
});
const makerOrdersResult = await getMakerOrders.json();
console.log("\nMaker Orders:");
if (makerOrdersResult.orders && makerOrdersResult.orders.length > 0) {
    const order = makerOrdersResult.orders[0];
    const filled = BigInt(order.filledAmtIn || 0);
    const total = BigInt(order.amtIn);
    const percent = (Number(filled) / Number(total) * 100).toFixed(2);
    console.log(`  Filled: ${filled.toString()} / ${total.toString()} (${percent}%)`);
    console.log(`  Status: ${order.status === 0 ? 'Active' : order.status === 2 ? 'Fully Filled' : 'Other'}`);
    console.log(`  Expected: 300000000000000000000 / 600000000000000000000 (50.00%)`);
} else {
    console.log("  No orders found or order fully filled and removed");
}

const getTakerOrders = await fetch(`http://localhost:11223/orders/${taker.account.address}`, {
    method: "GET"
});
const takerOrdersResult = await getTakerOrders.json();
console.log("\nTaker Orders (BOTTLENECK):");
if (takerOrdersResult.orders && takerOrdersResult.orders.length > 0) {
    const order = takerOrdersResult.orders[0];
    const filled = BigInt(order.filledAmtIn || 0);
    const total = BigInt(order.amtIn);
    const percent = (Number(filled) / Number(total) * 100).toFixed(2);
    console.log(`  Filled: ${filled.toString()} / ${total.toString()} (${percent}%)`);
    console.log(`  Status: ${order.status === 0 ? 'Active' : order.status === 2 ? 'Fully Filled' : 'Other'}`);
} else {
    console.log("  ✓ Order fully filled and removed from book (Expected: 100% filled)");
}

const getTaker2Orders = await fetch(`http://localhost:11223/orders/${taker2.account.address}`, {
    method: "GET"
});
const taker2OrdersResult = await getTaker2Orders.json();
console.log("\nTaker2 Orders:");
if (taker2OrdersResult.orders && taker2OrdersResult.orders.length > 0) {
    const order = taker2OrdersResult.orders[0];
    const filled = BigInt(order.filledAmtIn || 0);
    const total = BigInt(order.amtIn);
    const percent = (Number(filled) / Number(total) * 100).toFixed(2);
    console.log(`  Filled: ${filled.toString()} / ${total.toString()} (${percent}%)`);
    console.log(`  Status: ${order.status === 0 ? 'Active' : order.status === 2 ? 'Fully Filled' : 'Other'}`);
    console.log(`  Expected: 100000000000000000000 / 150000000000000000000 (66.67%)`);
} else {
    console.log("  No orders found or order fully filled and removed");
}

console.log("\n=== VERIFICATION ===");
console.log("Check backend logs for:");
console.log("1. Ring detection with 3 orders");
console.log("2. Bottleneck calculation = 200 DC");
console.log("3. Partial fills:");
console.log("   - Order 1 (Maker): 300 SC filled (50%)");
console.log("   - Order 2 (Taker): 200 DC filled (100%) - REMOVED");
console.log("   - Order 3 (Taker2): 100 WETH filled (66.67%)");
console.log("\n✓ Maker and Taker2 should remain in order book with updated filled amounts");
console.log("✓ Taker's order should be completely removed (100% filled)");
