import { network } from "hardhat";
import Token from "../artifacts/contracts/SmashCoin.sol/SmashCoin.json";

const { viem } = await network.connect({
  network: "localhost",
  chainType: "ethereum",
});

console.log("Getting balance")

const recipient = "0x1234567890123456789012345678901234567890";
const tokenAddr = "0x5FbDB2315678afecb367f032d93F642f64180aa3";
const publicClient = await viem.getPublicClient();
const accounts = await viem.getWalletClients();

const balance = await publicClient.readContract({
    address: tokenAddr,
    abi: Token.abi,
    functionName: "balanceOf",
    args: [recipient],
});

console.log(`Balance (raw): ${balance}`);

for (const acc of accounts) {
    const balance = await publicClient.readContract({
        address: tokenAddr,
        abi: Token.abi,
        functionName: "balanceOf",
        args: [acc.account.address],
    });
    console.log(`${acc.account.address}: ${balance}`);
}
