import assert from "assert";
import { network } from "hardhat";

describe("DEX", async function () {
  const { viem } = await network.connect();
  const [deployer, maker, taker] = await viem.getWalletClients();

  it("Exchange execute a full-fill swap EIP-712 order signatures", async function () {
    // ---------------------------
    // Deploy test tokens
    const toMint = 1_000_000n * 10n ** 18n;
    const tokenA = await viem.deployContract("Token1", [toMint]);
    const tokenB = await viem.deployContract("Token2", [toMint]);
    // Deploy registry
    const registry = await viem.deployContract("TokenRegistry");
    await registry.write.addToken([tokenA.address]);
    await registry.write.addToken([tokenB.address]);
    // Mint balances to maker/taker
    await tokenA.write.transfer([maker.account.address, 1_000n * 10n ** 18n]);
    await tokenB.write.transfer([taker.account.address, 1_000n * 10n ** 18n]);
    // Deploy Exchange
    const dex = await viem.deployContract("Exchange", [registry.address]);
    await tokenA.write.approve([dex.address, 1_000n * 10n ** 18n], { account: maker.account.address });
    await tokenB.write.approve([dex.address, 1_000n * 10n ** 18n], { account: taker.account.address });
    const chainId = await deployer.getChainId();
    // ---------------------------
    // Orders
    const makerOrder = {
      createdBy: maker.account.address,
      symbolIn: "SC",
      symbolOut: "DC",
      amtIn: 100n * 10n ** 18n,
      amtOut: 200n * 10n ** 18n,
      nonce: 1n,
    };
    const takerOrder = {
      createdBy: taker.account.address,
      symbolIn: "DC",
      symbolOut: "SC",
      amtIn: 200n * 10n ** 18n,
      amtOut: 100n * 10n ** 18n,
      nonce: 1n,
    };
    // ---------------------------
    // Helper: EIP-712 signing for orders
    async function signOrder(order: any, signer: any) {
      const domain = {
        name: "SmashDEX",
        version: "1",
        chainId,
        verifyingContract: dex.address,
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
      const signature = await signer.signTypedData({ domain, types, primaryType: "Order", message: order });
      return signature as `0x${string}`;
    }
    const makerSignature = await signOrder(makerOrder, maker);
    const takerSignature = await signOrder(takerOrder, taker);
    // ---------------------------
    // Execute swap using SwapInfo (full fill)
    const fillAmtIn = 100n * 10n ** 18n; // Fill entire maker order
    await dex.write.executeOrder([
      {
        order: makerOrder,
        signature: makerSignature,
      },
      {
        order: takerOrder,
        signature: takerSignature,
      },
      fillAmtIn,
    ]);
    // ---------------------------
    // Check balances
    const makerA = await tokenA.read.balanceOf([maker.account.address]);
    const makerB = await tokenB.read.balanceOf([maker.account.address]);
    const takerA = await tokenA.read.balanceOf([taker.account.address]);
    const takerB = await tokenB.read.balanceOf([taker.account.address]);
    assert.equal(makerA, 900n * 10n ** 18n, "Maker tokenA should decrease by 100");
    assert.equal(makerB, 200n * 10n ** 18n, "Maker tokenB should increase by 200");
    assert.equal(takerA, 100n * 10n ** 18n, "Taker tokenA should increase by 100");
    assert.equal(takerB, 800n * 10n ** 18n, "Taker tokenB should decrease by 200");
  });

  it("Exchange execute a 50% partial fill", async function () {
    // ---------------------------
    // Deploy test tokens
    const toMint = 1_000_000n * 10n ** 18n;
    const tokenA = await viem.deployContract("Token1", [toMint]);
    const tokenB = await viem.deployContract("Token2", [toMint]);
    // Deploy registry
    const registry = await viem.deployContract("TokenRegistry");
    await registry.write.addToken([tokenA.address]);
    await registry.write.addToken([tokenB.address]);
    // Mint balances to maker/taker
    await tokenA.write.transfer([maker.account.address, 1_000n * 10n ** 18n]);
    await tokenB.write.transfer([taker.account.address, 1_000n * 10n ** 18n]);
    // Deploy Exchange
    const dex = await viem.deployContract("Exchange", [registry.address]);
    await tokenA.write.approve([dex.address, 1_000n * 10n ** 18n], { account: maker.account.address });
    await tokenB.write.approve([dex.address, 1_000n * 10n ** 18n], { account: taker.account.address });
    const chainId = await deployer.getChainId();
    // ---------------------------
    // Orders - Maker willing to trade 100 SC for 200 DC
    const makerOrder = {
      createdBy: maker.account.address,
      symbolIn: "SC",
      symbolOut: "DC",
      amtIn: 100n * 10n ** 18n,
      amtOut: 200n * 10n ** 18n,
      nonce: 2n,
    };
    // Taker has enough to match the full order
    const takerOrder = {
      createdBy: taker.account.address,
      symbolIn: "DC",
      symbolOut: "SC",
      amtIn: 200n * 10n ** 18n,
      amtOut: 100n * 10n ** 18n,
      nonce: 2n,
    };
    // ---------------------------
    // Helper: EIP-712 signing for orders
    async function signOrder(order: any, signer: any) {
      const domain = {
        name: "SmashDEX",
        version: "1",
        chainId,
        verifyingContract: dex.address,
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
      const signature = await signer.signTypedData({ domain, types, primaryType: "Order", message: order });
      return signature as `0x${string}`;
    }
    const makerSignature = await signOrder(makerOrder, maker);
    const takerSignature = await signOrder(takerOrder, taker);
    // ---------------------------
    // Execute 50% partial fill
    const fillAmtIn = 50n * 10n ** 18n; // Fill only 50 SC (50% of maker's order)
    await dex.write.executeOrder([
      {
        order: makerOrder,
        signature: makerSignature,
      },
      {
        order: takerOrder,
        signature: takerSignature,
      },
      fillAmtIn,
    ]);
    // ---------------------------
    // Check balances after 50% fill
    const makerA = await tokenA.read.balanceOf([maker.account.address]);
    const makerB = await tokenB.read.balanceOf([maker.account.address]);
    const takerA = await tokenA.read.balanceOf([taker.account.address]);
    const takerB = await tokenB.read.balanceOf([taker.account.address]);
    
    // Maker gives 50 SC, receives 100 DC
    assert.equal(makerA, 950n * 10n ** 18n, "Maker tokenA should decrease by 50");
    assert.equal(makerB, 100n * 10n ** 18n, "Maker tokenB should increase by 100");
    
    // Taker gives 100 DC, receives 50 SC
    assert.equal(takerA, 50n * 10n ** 18n, "Taker tokenA should increase by 50");
    assert.equal(takerB, 900n * 10n ** 18n, "Taker tokenB should decrease by 100");

    // ---------------------------
    // Check filled amounts
    const makerFilled = await dex.read.filledOrdersAmtIn([maker.account.address, 2n]);
    const takerFilled = await dex.read.filledOrdersAmtIn([taker.account.address, 2n]);
    
    assert.equal(makerFilled, 50n * 10n ** 18n, "Maker should have 50 filled");
    assert.equal(takerFilled, 100n * 10n ** 18n, "Taker should have 100 filled");

    // ---------------------------
    // Check remaining amounts
    const makerRemaining = await dex.read.getRemainingAmt([
      maker.account.address, 
      2n, 
      100n * 10n ** 18n
    ]);
    const takerRemaining = await dex.read.getRemainingAmt([
      taker.account.address, 
      2n, 
      200n * 10n ** 18n
    ]);
    
    assert.equal(makerRemaining, 50n * 10n ** 18n, "Maker should have 50 remaining");
    assert.equal(takerRemaining, 100n * 10n ** 18n, "Taker should have 100 remaining");
  });

  it("Exchange execute multiple partial fills on the same order", async function () {
    // ---------------------------
    // Deploy test tokens
    const toMint = 1_000_000n * 10n ** 18n;
    const tokenA = await viem.deployContract("Token1", [toMint]);
    const tokenB = await viem.deployContract("Token2", [toMint]);
    // Deploy registry
    const registry = await viem.deployContract("TokenRegistry");
    await registry.write.addToken([tokenA.address]);
    await registry.write.addToken([tokenB.address]);
    // Mint balances to maker/taker
    await tokenA.write.transfer([maker.account.address, 1_000n * 10n ** 18n]);
    await tokenB.write.transfer([taker.account.address, 1_000n * 10n ** 18n]);
    // Deploy Exchange
    const dex = await viem.deployContract("Exchange", [registry.address]);
    await tokenA.write.approve([dex.address, 1_000n * 10n ** 18n], { account: maker.account.address });
    await tokenB.write.approve([dex.address, 1_000n * 10n ** 18n], { account: taker.account.address });
    const chainId = await deployer.getChainId();
    // ---------------------------
    // Helper: EIP-712 signing for orders
    async function signOrder(order: any, signer: any) {
      const domain = {
        name: "SmashDEX",
        version: "1",
        chainId,
        verifyingContract: dex.address,
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
      const signature = await signer.signTypedData({ domain, types, primaryType: "Order", message: order });
      return signature as `0x${string}`;
    }

    // ---------------------------
    // Maker's large order - willing to trade 500 SC for 1000 DC
    const makerOrder = {
      createdBy: maker.account.address,
      symbolIn: "SC",
      symbolOut: "DC",
      amtIn: 500n * 10n ** 18n,
      amtOut: 1000n * 10n ** 18n,
      nonce: 3n,
    };
    const makerSignature = await signOrder(makerOrder, maker);

    // ---------------------------
    // First partial fill - 40% (200 SC)
    const takerOrder1 = {
      createdBy: taker.account.address,
      symbolIn: "DC",
      symbolOut: "SC",
      amtIn: 400n * 10n ** 18n,
      amtOut: 200n * 10n ** 18n,
      nonce: 3n,
    };
    const takerSignature1 = await signOrder(takerOrder1, taker);

    await dex.write.executeOrder([
      { order: makerOrder, signature: makerSignature },
      { order: takerOrder1, signature: takerSignature1 },
      200n * 10n ** 18n, // Fill 200 SC (40% of maker's order)
    ]);

    // Check after first fill
    let makerFilled = await dex.read.filledOrdersAmtIn([maker.account.address, 3n]);
    assert.equal(makerFilled, 200n * 10n ** 18n, "After first fill: Maker should have 200 filled");

    // ---------------------------
    // Second partial fill - another 30% (150 SC)
    const takerOrder2 = {
      createdBy: taker.account.address,
      symbolIn: "DC",
      symbolOut: "SC",
      amtIn: 300n * 10n ** 18n,
      amtOut: 150n * 10n ** 18n,
      nonce: 4n, // Different nonce for new taker order
    };
    const takerSignature2 = await signOrder(takerOrder2, taker);

    await dex.write.executeOrder([
      { order: makerOrder, signature: makerSignature },
      { order: takerOrder2, signature: takerSignature2 },
      150n * 10n ** 18n, // Fill another 150 SC
    ]);

    // Check after second fill
    makerFilled = await dex.read.filledOrdersAmtIn([maker.account.address, 3n]);
    assert.equal(makerFilled, 350n * 10n ** 18n, "After second fill: Maker should have 350 filled total");

    // Check remaining
    const makerRemaining = await dex.read.getRemainingAmt([
      maker.account.address,
      3n,
      500n * 10n ** 18n,
    ]);
    assert.equal(makerRemaining, 150n * 10n ** 18n, "Maker should have 150 SC remaining");

    // ---------------------------
    // Check final balances
    const makerA = await tokenA.read.balanceOf([maker.account.address]);
    const makerB = await tokenB.read.balanceOf([maker.account.address]);

    // Maker gave 350 SC total, received 700 DC total
    assert.equal(makerA, 650n * 10n ** 18n, "Maker tokenA should decrease by 350");
    assert.equal(makerB, 700n * 10n ** 18n, "Maker tokenB should increase by 700");
  });

  it("Should reject partial fill that exceeds remaining order amount", async function () {
    // ---------------------------
    // Deploy test tokens
    const toMint = 1_000_000n * 10n ** 18n;
    const tokenA = await viem.deployContract("Token1", [toMint]);
    const tokenB = await viem.deployContract("Token2", [toMint]);
    // Deploy registry
    const registry = await viem.deployContract("TokenRegistry");
    await registry.write.addToken([tokenA.address]);
    await registry.write.addToken([tokenB.address]);
    // Mint balances to maker/taker
    await tokenA.write.transfer([maker.account.address, 1_000n * 10n ** 18n]);
    await tokenB.write.transfer([taker.account.address, 1_000n * 10n ** 18n]);
    // Deploy Exchange
    const dex = await viem.deployContract("Exchange", [registry.address]);
    await tokenA.write.approve([dex.address, 1_000n * 10n ** 18n], { account: maker.account.address });
    await tokenB.write.approve([dex.address, 1_000n * 10n ** 18n], { account: taker.account.address });
    const chainId = await deployer.getChainId();
    // ---------------------------
    // Helper: EIP-712 signing for orders
    async function signOrder(order: any, signer: any) {
      const domain = {
        name: "SmashDEX",
        version: "1",
        chainId,
        verifyingContract: dex.address,
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
      const signature = await signer.signTypedData({ domain, types, primaryType: "Order", message: order });
      return signature as `0x${string}`;
    }

    // Small maker order
    const makerOrder = {
      createdBy: maker.account.address,
      symbolIn: "SC",
      symbolOut: "DC",
      amtIn: 100n * 10n ** 18n,
      amtOut: 200n * 10n ** 18n,
      nonce: 5n,
    };

    const takerOrder = {
      createdBy: taker.account.address,
      symbolIn: "DC",
      symbolOut: "SC",
      amtIn: 200n * 10n ** 18n,
      amtOut: 100n * 10n ** 18n,
      nonce: 5n,
    };

    const makerSignature = await signOrder(makerOrder, maker);
    const takerSignature = await signOrder(takerOrder, taker);

    // Try to fill MORE than the order size
    try {
      await dex.write.executeOrder([
        { order: makerOrder, signature: makerSignature },
        { order: takerOrder, signature: takerSignature },
        150n * 10n ** 18n, // Trying to fill 150 when order is only 100
      ]);
      assert.fail("Should have thrown an error");
    } catch (error: any) {
      assert.ok(
        error.message.includes("Fill amount exceeds maker's remaining order"),
        "Should reject overfill"
      );
    }
  });
});

function describe(name: string, fn: Function) {
  console.log(`\nSuite: ${name}`);
  fn();
}

function it(name: string, fn: Function) {
  Promise.resolve(fn())
    .then(() => console.log(`  ✔ ${name}`))
    .catch((err) => console.error(`  ✖ ${name}\n`, err));
}
