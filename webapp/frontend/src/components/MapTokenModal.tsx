import React, { useState, useEffect } from "react";

interface Token {
  address: string;
  symbol: string;
  name: string;
}

export default function MapTokenModal({
  isOpen,
  onClose,
  initialSymbol,
  onSave,
}: {
  isOpen: boolean;
  onClose: () => void;
  initialSymbol?: string | null;
  onSave: (t: Token) => void;
}) {
  const [addr, setAddr] = useState("");
  const [symbol, setSymbol] = useState(initialSymbol || "");
  const [name, setName] = useState("");
  const [noAddress, setNoAddress] = useState(false);

  useEffect(() => {
    setSymbol(initialSymbol || "");
  }, [initialSymbol]);

  if (!isOpen) return null;

  function handleSave() {
    const a = addr.trim();
    const s = (symbol || "").trim();
    const n = (name || "").trim();
    if (!s) {
      alert("Token symbol is required");
      return;
    }
    if (!noAddress) {
      if (!a || !a.startsWith("0x") || a.length < 10) {
        alert("Please paste a valid token contract address starting with 0x or choose 'I don't have the address' to save as unmapped.");
        return;
      }
    }
    const token = { address: a, symbol: s, name: n || s };
    // reset local state
    setAddr("");
    setSymbol("");
    setName("");
    onSave(token);
    onClose();
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
      <div className="bg-zinc-900 text-zinc-100 rounded-lg p-6 w-full max-w-md">
        <h3 className="text-lg font-semibold mb-2">Add custom token</h3>
        <p className="text-sm text-zinc-400 mb-4">Add a custom token so it appears in the pair selector.</p>

        <label className="text-xs text-zinc-400">Symbol</label>
        <input
          value={symbol}
          onChange={(e) => setSymbol(e.target.value)}
          placeholder="e.g. DEX"
          className="w-full p-2 bg-zinc-800 rounded mb-3"
        />

        <label className="text-xs text-zinc-400">Name (optional)</label>
        <input
          value={name}
          onChange={(e) => setName(e.target.value)}
          placeholder="Token name"
          className="w-full p-2 bg-zinc-800 rounded mb-3"
        />

        <label className="text-xs text-zinc-400">Contract address</label>
        <input
          value={addr}
          onChange={(e) => setAddr(e.target.value)}
          placeholder="0x... token contract address"
          disabled={noAddress}
          className="w-full p-2 bg-zinc-800 rounded mb-2 font-mono"
        />

        <div className="flex items-center gap-2 text-sm text-zinc-400 mb-4">
          <input id="noaddr" type="checkbox" checked={noAddress} onChange={() => setNoAddress(!noAddress)} />
          <label htmlFor="noaddr">I don't have the contract address (save as unmapped)</label>
        </div>

        <div className="flex justify-end gap-2">
          <button onClick={onClose} className="px-3 py-1 rounded bg-zinc-700">Cancel</button>
          <button onClick={handleSave} className="px-3 py-1 rounded bg-indigo-600 hover:bg-indigo-500">Save</button>
        </div>
      </div>
    </div>
  );
}
