import { motion, AnimatePresence } from "framer-motion";
import type { Order } from "../App";
import { ethers } from "ethers";

export default function TradeHistory({ trades }: { trades: Order[] }) {
  function status(status: string) {
    if (status == "3" || status == "Completed") {
      return "Completed"
    } else if (status == "4" || status == "Cancelled") {
      return "Cancelled"
    } else if (status == "5" || status == "PartialFill") {
      return "PartialFill"
    }
  }
  return (
    <div className="bg-zinc-900 p-5 rounded-2xl h-[22rem] flex flex-col shadow-lg border border-zinc-800">
      <h2 className="font-semibold text-lg mb-3 text-indigo-400">Order History</h2>

      {/* Fixed header (no sticky) */}
      <div className="grid grid-cols-6 text-xs text-zinc-400 px-3 py-1 border-b border-zinc-800 shrink-0">
        <div className="text-left w-[84px] shrink-0">Nonce</div>
        <div className="text-left">Base Token</div>
        <div className="text-left">Quote Token</div>
        <div className="text-center">Base In</div>
        <div className="text-center">Quote Out</div>
        <div className="text-center">Status</div>
      </div>

      {/* Scrolling rows area */}
      <div className="min-h-0 flex-1 overflow-auto">
        <div className="overflow-hidden">
        <AnimatePresence initial={false}>
          {trades.map((order, i) => {

            return (
              <motion.div
                key={`parent-${i}`} // use index
                initial={{ opacity: 0, y: -8 }}
                animate={{ opacity: 1, y: 0 }}
                exit={{ opacity: 0, y: -8 }}
                transition={{ duration: 0.3 }}
                className="grid grid-cols-6 text-sm px-3 py-1 border-b border-zinc-800"
              >
                <div className="truncate text-left">{order.nonce}</div>
                <div className="truncate text-center">{order.symbolIn}</div>
                <div className="truncate text-center">{order.symbolOut}</div>
                <div className="truncate text-center">
                  {ethers.formatEther(BigInt(order.amtIn))}
                  {order.filledAmtIn && order.filledAmtIn !== order.amtIn
                    ? ` (${ethers.formatEther(BigInt(order.filledAmtIn))})`
                    : ""}
                </div>
                <div className="truncate text-center">
                  {ethers.formatEther(BigInt(order.amtOut))}
                  {order.filledAmtIn && order.filledAmtIn !== order.amtIn
                    ? ` (${ethers.formatEther((BigInt(order.amtOut) * BigInt(order.filledAmtIn)) / BigInt(order.amtIn))})`
                    : ""}
                </div>
                <div className="truncate text-center">{status(order.status!)}</div>

                {/* Conditional order */}
                {/* {order.conditionalOrder && (
                  <motion.div
                    key={`cond-${i}`} // use same index + prefix
                    initial={{ opacity: 0, y: -8 }}
                    animate={{ opacity: 1, y: 0 }}
                    exit={{ opacity: 0, y: -8 }}
                    transition={{ duration: 0.3 }}
                    className="grid grid-cols-6 text-sm px-3 py-1 border-b border-zinc-800 text-zinc-400"
                  >
                    <div className="truncate text-left">{order.conditionalOrder.nonce}</div>
                    <div className="truncate text-center">{order.conditionalOrder.symbolIn}</div>
                    <div className="truncate text-center">{order.conditionalOrder.symbolOut}</div>
                    <div className="truncate text-center">{ethers.formatEther(BigInt(order.conditionalOrder.amtIn))}</div>
                    <div className="truncate text-center">{ethers.formatEther(BigInt(order.conditionalOrder.amtOut))}</div>
                    <div className="truncate text-center">Conditional</div>
                  </motion.div>
                )} */}
              </motion.div>
            );
          })}
        </AnimatePresence>
        </div>
      </div>
    </div>
  );
}
