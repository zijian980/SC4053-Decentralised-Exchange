import { useState } from "react";

export interface TokenQuantity {
  tokenIn: bigint; // base
  tokenOut: bigint; // quote
  hasCondition: boolean;
  conditionalType?: string;
  conditionalStopPrice?: bigint;
  conditionalLimitPrice?: bigint;
  conditionalTokenIn?: bigint;
  conditionalTokenOut?: bigint;
}

export default function OrderForm({
  onSubmit,
}: {
  onSubmit: (details: TokenQuantity) => void;
}) {
  const [price, setPrice] = useState("");
  const [tokenIn, setTokenIn] = useState(""); // Base
  const [tokenOut, setTokenOut] = useState(""); // Quote
  const [editingField, setEditingField] = useState<"tokenIn" | "tokenOut" | null>(null);

  const [isConditional, setIsConditional] = useState(false);
  const [conditionType, setConditionType] = useState<"STOP_LIMIT">("STOP_LIMIT");

  // Conditional Fields
  const [conditionalPrice, setConditionalPrice] = useState(""); // Stop Trigger
  const [conditionalLimitPrice, setConditionalLimitPrice] = useState(""); // Limit Price
  const [conditionalTokenIn, setConditionalTokenIn] = useState(""); // Base
  const [conditionalTokenOut, setConditionalTokenOut] = useState(""); // Quote
  const [conditionalEditingField, setConditionalEditingField] = useState<
    "tokenIn" | "tokenOut" | null
  >(null);

  // === BASIC ORDER LOGIC ===
  function handlePriceChange(e: React.ChangeEvent<HTMLInputElement>) {
    setPrice(e.target.value);
  }

  function handleTokenInChange(e: React.ChangeEvent<HTMLInputElement>) {
    setTokenIn(e.target.value);
    setEditingField("tokenIn");
  }

  function handleTokenOutChange(e: React.ChangeEvent<HTMLInputElement>) {
    setTokenOut(e.target.value);
    setEditingField("tokenOut");
  }

  function handlePriceBlur() {
    const p = parseFloat(price);
    if (isNaN(p) || p <= 0) return;

    if (editingField === "tokenIn") {
      // base → quote
      const base = parseFloat(tokenIn || "0");
      if (!isNaN(base)) setTokenOut((base * p).toFixed(6));
    } else if (editingField === "tokenOut") {
      // quote → base
      const quote = parseFloat(tokenOut || "0");
      if (!isNaN(quote)) setTokenIn((quote / p).toFixed(6));
    }
  }

  function handleTokenInBlur() {
    const p = parseFloat(price);
    const base = parseFloat(tokenIn || "0");
    const quote = parseFloat(tokenOut || "0");

    if (!isNaN(p) && p > 0) {
      setTokenOut((base * p).toFixed(6));
    } else if (base > 0 && quote > 0) {
      const derivedPrice = quote / base;
      if (isFinite(derivedPrice)) setPrice(derivedPrice.toFixed(6));
    }
  }

  function handleTokenOutBlur() {
    const p = parseFloat(price);
    const quote = parseFloat(tokenOut || "0");
    const base = parseFloat(tokenIn || "0");

    if (!isNaN(p) && p > 0) {
      setTokenIn((quote / p).toFixed(6));
    } else if (base > 0 && quote > 0) {
      const derivedPrice = quote / base;
      if (isFinite(derivedPrice)) setPrice(derivedPrice.toFixed(6));
    }
  }

  // === CONDITIONAL ORDER LOGIC ===
  function handleConditionalLimitPriceChange(e: React.ChangeEvent<HTMLInputElement>) {
    setConditionalLimitPrice(e.target.value);
  }

  function handleConditionalTokenInChange(e: React.ChangeEvent<HTMLInputElement>) {
    setConditionalTokenIn(e.target.value);
    setConditionalEditingField("tokenIn");
  }

  function handleConditionalTokenOutChange(e: React.ChangeEvent<HTMLInputElement>) {
    setConditionalTokenOut(e.target.value);
    setConditionalEditingField("tokenOut");
  }

  function handleConditionalLimitPriceBlur() {
    const p = parseFloat(conditionalLimitPrice);
    if (isNaN(p) || p <= 0) return;

    if (conditionalEditingField === "tokenIn") {
      const base = parseFloat(conditionalTokenIn || "0");
      if (!isNaN(base)) setConditionalTokenOut((base * p).toFixed(6));
    } else if (conditionalEditingField === "tokenOut") {
      const quote = parseFloat(conditionalTokenOut || "0");
      if (!isNaN(quote)) setConditionalTokenIn((quote / p).toFixed(6));
    }
  }

  function handleConditionalTokenInBlur() {
    const p = parseFloat(conditionalLimitPrice);
    const base = parseFloat(conditionalTokenIn || "0");
    const quote = parseFloat(conditionalTokenOut || "0");

    if (!isNaN(p) && p > 0) {
      setConditionalTokenOut((base * p).toFixed(6));
    } else if (base > 0 && quote > 0) {
      const derivedPrice = quote / base;
      if (isFinite(derivedPrice)) setConditionalLimitPrice(derivedPrice.toFixed(6));
    }
  }

  function handleConditionalTokenOutBlur() {
    const p = parseFloat(conditionalLimitPrice);
    const quote = parseFloat(conditionalTokenOut || "0");
    const base = parseFloat(conditionalTokenIn || "0");

    if (!isNaN(p) && p > 0) {
      setConditionalTokenIn((quote / p).toFixed(6));
    } else if (base > 0 && quote > 0) {
      const derivedPrice = quote / base;
      if (isFinite(derivedPrice)) setConditionalLimitPrice(derivedPrice.toFixed(6));
    }
  }

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!price || !tokenIn || !tokenOut) return alert("Fill all fields");

    const tokenDecimals = 18;
    const baseDec = Number(tokenIn);
    const quoteDec = Number(tokenOut);

    if (isNaN(baseDec) || isNaN(quoteDec)) return alert("Invalid amounts");

    const amtIn = BigInt(Math.floor(baseDec * 10 ** tokenDecimals)); // base
    const amtOut = BigInt(Math.floor(quoteDec * 10 ** tokenDecimals)); // quote

    let conditionalAmtIn: bigint | undefined;
    let conditionalAmtOut: bigint | undefined;
    let condStopPrice: bigint | undefined;
    let condLimitPrice: bigint | undefined;

    if (isConditional) {
      if (
        !conditionalPrice ||
        !conditionalLimitPrice ||
        !conditionalTokenIn ||
        !conditionalTokenOut
      ) {
        return alert("Fill all conditional order fields");
      }

      const stop = Number(conditionalPrice);
      const limit = Number(conditionalLimitPrice);
      const baseC = Number(conditionalTokenIn);
      const quoteC = Number(conditionalTokenOut);

      if (isNaN(stop) || isNaN(baseC) || isNaN(quoteC)) {
        return alert("Invalid conditional values");
      }

    if (isNaN(stop) || isNaN(limit) || isNaN(baseC) || isNaN(quoteC)) {
      return alert("Invalid conditional values");
    }

    condStopPrice = BigInt(Math.floor(stop * 10 ** tokenDecimals));
    condLimitPrice = BigInt(Math.floor(limit * 10 ** tokenDecimals));
    conditionalAmtIn = BigInt(Math.floor(baseC * 10 ** tokenDecimals));
    conditionalAmtOut = BigInt(Math.floor(quoteC * 10 ** tokenDecimals));
    }

    onSubmit({
      tokenIn: amtIn,
      tokenOut: amtOut,
      hasCondition: isConditional,
      conditionalType: isConditional ? conditionType : undefined,
      conditionalStopPrice: condStopPrice,
      conditionalLimitPrice: condLimitPrice,
      conditionalTokenIn: conditionalAmtIn,
      conditionalTokenOut: conditionalAmtOut,
    });

    // reset
    setPrice("");
    setTokenIn("");
    setTokenOut("");
    setEditingField(null);
    setIsConditional(false);
    setConditionType("STOP_LIMIT");
    setConditionalPrice("");
    setConditionalLimitPrice("");
    setConditionalTokenIn("");
    setConditionalTokenOut("");
    setConditionalEditingField(null);
  }

  return (
    <div className="bg-zinc-900 rounded-2xl p-5 shadow-lg border border-zinc-800">
      <h2 className="font-semibold text-lg mb-3 text-indigo-400">Place Order</h2>
      <form onSubmit={handleSubmit} className="space-y-3">
        {/* Regular Order */}
        <input
          type="number"
          step="0.01"
          placeholder="Price (quote per base)"
          value={price}
          onChange={handlePriceChange}
          onBlur={handlePriceBlur}
          className="w-full p-2 bg-zinc-800 rounded text-sm"
        />

        <input
          type="number"
          step="0.01"
          placeholder="Base Quantity (token in)"
          value={tokenIn}
          onChange={handleTokenInChange}
          onBlur={handleTokenInBlur}
          className="w-full p-2 bg-zinc-800 rounded text-sm"
        />

        <input
          type="number"
          step="0.01"
          placeholder="Quote Quantity (token out)"
          value={tokenOut}
          onChange={handleTokenOutChange}
          onBlur={handleTokenOutBlur}
          className="w-full p-2 bg-zinc-800 rounded text-sm"
        />

        {/* Conditional Order Toggle */}
        <div className="flex items-center justify-between py-2 border-t border-zinc-800 mt-2">
          <span className="text-sm font-medium text-zinc-300">Conditional Order</span>
          <label className="relative inline-flex items-center cursor-pointer">
            <input
              type="checkbox"
              checked={isConditional}
              onChange={() => setIsConditional(!isConditional)}
              className="sr-only peer"
            />
            <div className="w-10 h-5 bg-zinc-700 peer-focus:outline-none rounded-full peer peer-checked:bg-emerald-600 after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border after:rounded-full after:h-4 after:w-4 after:transition-all peer-checked:after:translate-x-5"></div>
          </label>
        </div>

        {isConditional && (
          <div className="space-y-3 mt-3 p-3 border border-zinc-800 rounded-xl bg-zinc-950">
            <label
              htmlFor="condition-type"
              className="block text-xs font-medium text-zinc-400 mb-1"
            >
              Condition Type
            </label>
            <select
              id="condition-type"
              value={conditionType}
              onChange={(e) => setConditionType(e.target.value as "STOP_LIMIT")}
              className="w-full p-2 bg-zinc-800 rounded text-sm"
            >
              <option value="STOP_LIMIT">Stop-Limit</option>
            </select>

            <div className="text-xs text-zinc-400">
              Triggers a limit order if the price reaches the Stop Price.
            </div>

            <input
              type="number"
              step="0.01"
              placeholder="Stop Price (trigger price)"
              value={conditionalPrice}
              onChange={(e) => setConditionalPrice(e.target.value)}
              className="w-full p-2 bg-zinc-800 rounded text-sm"
            />

            <input
              type="number"
              step="0.01"
              placeholder="Limit Price (order price)"
              value={conditionalLimitPrice}
              onChange={handleConditionalLimitPriceChange}
              onBlur={handleConditionalLimitPriceBlur}
              className="w-full p-2 bg-zinc-800 rounded text-sm"
            />

            <input
              type="number"
              step="0.01"
              placeholder="Conditional Base Quantity (token in)"
              value={conditionalTokenIn}
              onChange={handleConditionalTokenInChange}
              onBlur={handleConditionalTokenInBlur}
              className="w-full p-2 bg-zinc-800 rounded text-sm"
            />

            <input
              type="number"
              step="0.01"
              placeholder="Conditional Quote Quantity (token out)"
              value={conditionalTokenOut}
              onChange={handleConditionalTokenOutChange}
              onBlur={handleConditionalTokenOutBlur}
              className="w-full p-2 bg-zinc-800 rounded text-sm"
            />
          </div>
        )}

        <button
          type="submit"
          className="w-full py-2 rounded font-semibold bg-emerald-600 hover:bg-emerald-500 transition-colors"
        >
          Place Order
        </button>
      </form>
    </div>
  );
}
