import { ethers } from "ethers";
import ABI from "../artifacts/contracts/DoogeCoin.sol/DoogeCoin.json"

// Connect to a JSON-RPC endpoint
const provider = new ethers.JsonRpcProvider("http://127.0.0.1:8545");

// Example: get the latest block number
async function getBlockNumber() {
    const blockNumber = await provider.getBlockNumber();
    console.log("Latest Block Number:", blockNumber);
}

``

const tokenAddress = "0x663F3ad617193148711d28f5334eE4Ed07016602";
const owner = "0x15d34aaf54267db7d7c367839aaf71a00a2c6a65";
const spender = "0x9fE46736679d2D9a65F0992F2272dE9f3c7fa6e0";

const tokenContract = new ethers.Contract(tokenAddress, ABI.abi, provider);

async function getAllowance() {
    const allowance = await tokenContract.allowance(owner, spender);
    console.log("Allowance:", ethers.formatUnits(allowance, 18)); // 18 decimals typical
}

getAllowance();