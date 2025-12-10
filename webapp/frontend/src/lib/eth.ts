import { ethers } from "ethers";

// read-only provider to your local Hardhat node
export const readProvider = new ethers.JsonRpcProvider(import.meta.env.VITE_RPC_URL);
