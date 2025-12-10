import assert from "assert";
import { network } from "hardhat";

describe("ERC20 Token", async function () {
  const { viem } = await network.connect();
  const [deployer, maker] = await viem.getWalletClients(); // 1st EOA is used to deploy, 2nd and 3rd are maker and taker

  it("Token Creation", async () => {
    const toMint = 1000000n * 10n ** 18n;
    const tokenA = await viem.deployContract("Token1", [toMint]);

    const mintedValue = await tokenA.read.balanceOf([deployer.account.address]);

    assert.notEqual(tokenA.address, "0x0000000000000000000000000000000000000000", "Token not deployed");
    assert.equal(mintedValue, toMint, "Minted value not same as toMint value");
  });

  it("Token Transfer", async () => {
    const toMint = 1000000n * 10n ** 18n;
    const tokenA = await viem.deployContract("Token1", [toMint]);
    
    const toTransfer = 1_000n * 10n ** 18n;
    await tokenA.write.transfer([maker.account.address, toTransfer]);

    const makerBalance = await tokenA.read.balanceOf([maker.account.address]);

    assert.equal(makerBalance, toTransfer, "Receiver's balance should be the same as toTransfer value");
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

