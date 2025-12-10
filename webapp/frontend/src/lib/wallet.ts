import { ethers } from "ethers";

export async function connectWallet(): Promise<{
  address: string; chainId: number; provider: ethers.BrowserProvider;
}> {
  if (!(window as any).ethereum) throw new Error("MetaMask not found");
  const provider = new ethers.BrowserProvider((window as any).ethereum);
  await provider.send("eth_requestAccounts", []);
  const signer = await provider.getSigner();
  const network = await provider.getNetwork();
  return { address: await signer.getAddress(), chainId: Number(network.chainId), provider };
}

export async function switchOrAddNetwork(networkParams) {
  const { chainId } = networkParams;

  if (!window.ethereum) {
    alert("Please install MetaMask first.");
    return;
  }

  try {
    // Try switching first
    await window.ethereum.request({
      method: "wallet_switchEthereumChain",
      params: [{ chainId }],
    });
    console.log("✅ Switched to existing network");
  } catch (switchError) {
    // Error code 4902 means the chain has not been added
    if (switchError.code === 4902) {
      try {
        await window.ethereum.request({
          method: "wallet_addEthereumChain",
          params: [networkParams],
        });
        console.log("✅ Network added and switched");
      } catch (addError) {
        console.error("❌ Failed to add network", addError);
      }
    } else {
      console.error("❌ Failed to switch network", switchError);
    }
  }
}