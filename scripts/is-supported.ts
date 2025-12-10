import { network } from "hardhat";
import TokenRegistry from "../artifacts/contracts/TokenRegistry.sol/TokenRegistry.json";

const { viem } = await network.connect({
  network: "localhost",
  chainType: "ethereum",
});
const publicClient = await viem.getPublicClient();

const check = await publicClient.readContract({
    address: "0x9fe46736679d2d9a65f0992f2272de9f3c7fa6e0",
    abi: TokenRegistry.abi,
    functionName: "isSupported",
    args: ["0x5FbDB2315678afecb367f032d93F642f64180aa3"],
});

let gas = await publicClient.estimateContractGas({
    address: "0x9fe46736679d2d9a65f0992f2272de9f3c7fa6e0",
    abi: TokenRegistry.abi,
    functionName: "isSupported",
    args: ["0x5FbDB2315678afecb367f032d93F642f64180aa3"],
});
console.log(`Gas estimate ${gas}`)
console.log(`Supported by exchange check: ${check}`);
