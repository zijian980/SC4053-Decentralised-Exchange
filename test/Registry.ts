import assert from "assert";
import { network } from "hardhat";

describe("Token Registry", async function () {
  const { viem } = await network.connect();

  it("Registry Token Supported", async () => {
    // Deploy test tokens
    const toMint = 1000000n * 10n ** 18n;
    const tokenA = await viem.deployContract("Token1", [toMint]);
    const tokenB = await viem.deployContract("Token2", [toMint]);
    const registry = await viem.deployContract("TokenRegistry");
    await registry.write.addToken([tokenA.address]);
    await registry.write.addToken([tokenB.address]);

    // Check support of tokenA and tokenB
    const supportTokenA = await registry.read.isSupported([tokenA.address]);
    const supportTokenB = await registry.read.isSupported([tokenB.address]);


    // Assertions
    assert.equal(supportTokenA, true, "Token A should be supported");
    assert.equal(supportTokenB, true, "Token B should be supported");
  });

  it("Registry Symbol to Address conversion", async () => {
    // Deploy test tokens
    const toMint = 1000000n * 10n ** 18n;
    const tokenA = await viem.deployContract("Token1", [toMint]);
    const tokenB = await viem.deployContract("Token2", [toMint]);
    const registry = await viem.deployContract("TokenRegistry");
    await registry.write.addToken([tokenA.address]);
    await registry.write.addToken([tokenB.address]);
    
    // Check conversion of token symbol to token address
    const addrTokenA = await registry.read.getTokenAddress(["SC"]);
    const addrTokenB = await registry.read.getTokenAddress(["DC"]);

    assert.equal(addrTokenA.toLowerCase(), tokenA.address, "Retrieved address of token A should match token A");
    assert.equal(addrTokenB.toLowerCase(), tokenB.address, "Retrieved address of token B should match token B");
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

