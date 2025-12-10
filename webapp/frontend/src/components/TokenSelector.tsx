interface Token {
  address: string;
  name: string;
  symbol: string;
}

export default function TokenSelector({
  tokens,
  baseToken,
  quoteToken,
  setBaseToken,
  setQuoteToken,
  onSwap
}: {
  tokens: Token[];
  baseToken: string;
  quoteToken: string;
  setBaseToken: (s: string) => void;
  setQuoteToken: (s: string) => void;
  onSwap: () => void;
}) {
  const tokenList = Object.values(tokens); // convert Record to array

  function handleBaseChange(e: React.ChangeEvent<HTMLSelectElement>) {
    const selected = e.target.value;
    if (selected === quoteToken) {
      // Swap tokens
      setBaseToken(quoteToken);
      setQuoteToken(baseToken);
    } else {
      setBaseToken(selected);
    }
  }

  function handleQuoteChange(e: React.ChangeEvent<HTMLSelectElement>) {
    const selected = e.target.value;
    if (selected === baseToken) {
      // Swap tokens
      setBaseToken(quoteToken);
      setQuoteToken(baseToken);
    } else {
      setQuoteToken(selected);
    }
  }

  function swapTokens() {
    setBaseToken(quoteToken);
    setQuoteToken(baseToken);
    onSwap();
  }

  return (
    <div className="bg-zinc-900 p-3 rounded-xl">
      <div className="text-sm text-zinc-400 mb-2">Select pair</div>
      <div className="flex gap-2 items-center">
        <select
          value={baseToken}
          onChange={handleBaseChange}
          className="bg-zinc-800 p-2 rounded w-1/2 text-sm"
        >
          {tokenList.map((t) => (
            <option key={t.address || t.symbol} value={t.address} disabled={!t.address || !t.address.startsWith("0x")}>
              {t.symbol}{(!t.address || !t.address.startsWith("0x")) ? " (unmapped)" : ""}
            </option>
          ))}
        </select>

        <button
          onClick={swapTokens}
          className="bg-zinc-700 p-2 rounded text-sm text-white hover:bg-zinc-600 transition"
        >
          ⇄
        </button>

        <select
          value={quoteToken}
          onChange={handleQuoteChange}
          className="bg-zinc-800 p-2 rounded w-1/2 text-sm"
        >
          {tokenList.map((t) => (
            <option key={t.address || t.symbol} value={t.address} disabled={!t.address || !t.address.startsWith("0x")}>
              {t.symbol}{(!t.address || !t.address.startsWith("0x")) ? " (unmapped)" : ""}
            </option>
          ))}
        </select>
      </div>
    </div>
  );
}
