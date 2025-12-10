import { motion, AnimatePresence } from "framer-motion";
import { formatUnits } from "ethers";

function fmtQty(wei: string | bigint) {
  const qtyStr = formatUnits(typeof wei === "bigint" ? wei : BigInt(wei), 18);
  const qtyNum = Number(qtyStr);
  const shown = qtyNum > 0 && qtyNum < 0.0001 ? 0.0001 : qtyNum;
  return shown.toLocaleString(undefined, { minimumFractionDigits: 4, maximumFractionDigits: 4 });
}

export type OrderBookLevel = { 
  id: string; 
  side: "buy" | "sell";
  price: number;
  quantity: string;
  orderCount: number; }; 
  
export type OrderBooks = { 
  symbolIn: string,
  baseSymbol: string,
  ask: OrderBookLevel[],
  bid: OrderBookLevel[],
}

export default function OrderBook({ orderbooks }: { orderbooks: OrderBooks }) {
  const bids = [...orderbooks.bid].sort((a, b) => b.price - a.price);
  const asks = [...orderbooks.ask].sort((a, b) => a.price - b.price);

  return (
    <div className="bg-zinc-900 p-5 rounded-2xl h-[22rem] flex flex-col shadow-lg border border-zinc-800">
      <h2 className="font-semibold text-lg mb-3 text-indigo-400">Order Book</h2>

      {/* Fixed header */}
      <div className="grid grid-cols-3 text-xs text-zinc-400 px-3 py-1 border-b border-zinc-800 shrink-0">
        <div>Price</div>
        <div>Quantity</div>
        <div>Order Count</div>
      </div>

      {/* Scrolling rows */}
      <div className="min-h-0 flex-1 overflow-auto">
        <div className="overflow-hidden">
          <AnimatePresence initial={false}>
            {bids.map((r) => (
              <motion.div
                key={`bid-${r.id}`}
                initial={{ opacity: 0, y: -8 }}
                animate={{ opacity: 1, y: 0 }}
                exit={{ opacity: 0, y: -8 }}
                transition={{ duration: 0.3 }}
                className="grid grid-cols-3 text-sm px-3 py-1 border-b border-zinc-800 text-emerald-400"
              >
                <div className="font-mono">{r.price.toFixed(2)}</div>
                <div className="font-mono">{fmtQty(r.quantity)}</div>
                <div className="font-mono">
                  {r.orderCount}
                </div>
              </motion.div>
            ))}

            {asks.map((r) => (
              <motion.div
                key={`ask-${r.id}`}
                initial={{ opacity: 0, y: -8 }}
                animate={{ opacity: 1, y: 0 }}
                exit={{ opacity: 0, y: -8 }}
                transition={{ duration: 0.3 }}
                className="grid grid-cols-3 text-sm px-3 py-1 border-b border-zinc-800 text-rose-400"
              >
                <div className="font-mono">{r.price.toFixed(2)}</div>
                <div className="font-mono">{fmtQty(r.quantity)}</div>
                <div className="font-mono">{r.orderCount}</div>
              </motion.div>
            ))}
          </AnimatePresence>
        </div>
      </div>
    </div>
  );
}
