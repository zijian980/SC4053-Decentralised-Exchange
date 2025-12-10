import React, { useState } from "react";
import GenericTokenJson from "../abi/GenericToken.json";
import { ethers } from "ethers";
import { addToken } from "../lib/api";

export default function IssueTokenModal({
  isOpen,
  onClose,
  onDeployed,
}: {
  isOpen: boolean;
  onClose: () => void;
  onDeployed: (token: { address: string; symbol: string; name: string }) => void;
}) {
  const [name, setName] = useState("");
  const [symbol, setSymbol] = useState("");
  const [supply, setSupply] = useState("1000");
  const [busy, setBusy] = useState(false);

  if (!isOpen) return null;

  const hasBytecode = (GenericTokenJson as any).bytecode && (GenericTokenJson as any).bytecode.length > 10;

  async function handleDeploy() {
    if (!symbol || !name) return alert("Fill name and symbol");
    if (!hasBytecode) {
      return alert(
        "GenericToken artifact missing or bytecode empty. Run `npx hardhat compile` at the repo root and copy the generated artifacts/contracts/GenericToken.sol/GenericToken.json into webapp/frontend/src/abi/GenericToken.json"
      );
    }

    try {
      setBusy(true);
      await (window as any).ethereum.request({ method: "eth_requestAccounts" });
      const provider = new ethers.BrowserProvider((window as any).ethereum);
      const signer = await provider.getSigner();

      const factory = new ethers.ContractFactory((GenericTokenJson as any).abi, (GenericTokenJson as any).bytecode, signer as any);
      const initialSupply = BigInt(Math.floor(Number(supply) * 10 ** 18));
      const contract = await factory.deploy(name, symbol, initialSupply);
      if ((contract as any).waitForDeployment) {
        await (contract as any).waitForDeployment();
      }
      const addr = typeof (contract as any).getAddress === "function" ? await (contract as any).getAddress() : (contract as any).target || (contract as any).address;
      console.log(`Name: ${name} Symbol: ${symbol} Address: ${addr}`)
      try {
        await addToken(name, symbol, addr)
        onDeployed({ address: addr, symbol, name });
        alert(`Token deployed at ${addr}`)
      } catch {
        alert(`Token add failed`)
      }
      onClose();
    } catch (e: any) {
      console.error(e);
      alert("Deploy failed: " + (e?.message || e));
    } finally {
      setBusy(false);
    }
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
      <div className="bg-zinc-900 text-zinc-100 rounded-lg p-6 w-full max-w-md">
        <h3 className="text-lg font-semibold mb-2">Issue new token</h3>
        <p className="text-sm text-zinc-400 mb-4">Deploy a new ERC20 token from your wallet (this will cost gas).</p>

        <input value={name} onChange={(e) => setName(e.target.value)} placeholder="Token name" className="w-full p-2 bg-zinc-800 rounded mb-3" />
        <input value={symbol} onChange={(e) => setSymbol(e.target.value)} placeholder="Symbol (e.g. DEX)" className="w-full p-2 bg-zinc-800 rounded mb-3" />
        <input value={supply} onChange={(e) => setSupply(e.target.value)} placeholder="Initial supply (tokens)" className="w-full p-2 bg-zinc-800 rounded mb-3" />

        {!hasBytecode && (
          <div className="text-xs text-yellow-300 mb-3">
            Artifact bytecode not found. To enable deployment run:
            <div className="font-mono text-xs mt-1">npx hardhat compile</div>
            Then copy <span className="font-mono">artifacts/contracts/GenericToken.sol/GenericToken.json</span> into
            <span className="font-mono"> webapp/frontend/src/abi/GenericToken.json</span>
          </div>
        )}

        <div className="flex justify-end gap-2">
          <button onClick={onClose} className="px-3 py-1 rounded bg-zinc-700">Cancel</button>
          <button onClick={handleDeploy} disabled={busy} className="px-3 py-1 rounded bg-indigo-600 hover:bg-indigo-500">{busy ? 'Deploying...' : 'Deploy'}</button>
        </div>
      </div>
    </div>
  );
}
