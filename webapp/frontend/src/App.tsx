import { useEffect, useState, useRef, useCallback } from "react";
import { ethers } from "ethers";
import { readProvider } from "./lib/eth";
import { connectWallet, switchOrAddNetwork } from "./lib/wallet";
import { getNextNonce, signOrder, submitLimitOrderToBackend, openOrderWs, getOrdersByAddress, cancelOrder, getOrderBookNormalized, normalizeOrderBook, EXCHANGE_ADDRESS, getPastHistory } from "./lib/api";
// import MapTokenModal from "./components/MapTokenModal";
import IssueTokenModal from "./components/IssueTokenModal";
import { motion, AnimatePresence } from "framer-motion";
import TokenAbi from "./abi/Token.json";
import TokenRegistryAbi from "./abi/TokenRegistry.json";
import OrderForm, {type TokenQuantity} from "./components/OrderForm";

// import components
import OrderBook, { type OrderBooks } from "./components/OrderBook";
// import TradeHistory from "./components/TradeHistory";
import TokenSelector from "./components/TokenSelector";
import TradeHistory from "./components/TradeHistory";

const TOKENABI = TokenAbi.abi;
const REGISTRYABI = TokenRegistryAbi.abi;
const registryAddr: string = import.meta.env.VITE_TOKENREGISTRY || "";
const backendUrl = import.meta.env.VITE_BACKEND_URL;

export type Order = {
  createdBy: string;
  symbolIn: string;
  symbolOut: string;
  amtIn: bigint;
  amtOut: bigint;
  nonce: bigint;
  limitPrice?: bigint;
  status?: string;
  filledAmtIn?: bigint;
  signature?: string;
  conditionalOrder?: {
    createdBy: string;
    symbolIn: string;
    symbolOut: string;
    amtIn: bigint;
    amtOut: bigint;
    nonce: bigint;
    signature: string;
  };
};

interface Token {
  address: string,
  name: string,
  symbol: string,
}

export default function App() {
  const [status, setStatus] = useState("Loading…");
  const [chainId, setChainId] = useState<number>(0);
  //const [balance0, setBalance0] = useState("-");

  // Wallet state
  const [acct, setAcct] = useState<string>("(not connected)");
  const wsRef = useRef<WebSocket | null>(null);
  const wsNotifyRef = useRef<WebSocket | null>(null);
  const wsOrderBookRef = useRef<WebSocket | null>(null);

  const [orders, setOrders] = useState<Order[]>([]);
  const [orderBook, setOrderBook] = useState<OrderBooks>({baseSymbol: "", symbolIn: "",ask: [], bid: []});
  const [userOrders, setUserOrders] = useState<Array<Order>>([]);
  const [pastOrders, setPastOrders] = useState<Array<Order>>([]);
  const [lastPrice, setLastPrice] = useState<string>("-");
  const [isSwapped, setIsSwapped] = useState<boolean>(false);


  // Available tokens for pair selection (populated from on-chain TokenRegistry when available)
  const [availableTokens, setAvailableTokens] = useState<Token[]>([]); // Symbol: Token

  // Selected pair state (base / quote)
  const [baseToken, setBaseToken] = useState<string>("");
  const [quoteToken, setQuoteToken] = useState<string>("");
  // User balances for selected pair (human readable)
  const [baseBalanceDisplay, setBaseBalanceDisplay] = useState<string>("-");
  const [quoteBalanceDisplay, setQuoteBalanceDisplay] = useState<string>("-");
  const [mapModalOpen, setMapModalOpen] = useState(false);
  const [mapTokenSymbol, setMapTokenSymbol] = useState<string | null>(null);
  const [issueModalOpen, setIssueModalOpen] = useState(false);

  // fetch and set user's balances for selected base/quote tokens (component-level)
  const fetchBalances = useCallback(async (addr: string) => {
    try {
      if (baseToken) {
        const bc = new ethers.Contract(baseToken, TOKENABI, readProvider);
        const dec = Number(await bc.decimals?.().then((x) => Number(x)).catch(() => 18));
        const bal = await bc.balanceOf(addr);
        setBaseBalanceDisplay(ethers.formatUnits(bal, dec));
      } else {
        setBaseBalanceDisplay("-");
      }
      if (quoteToken) {
        const bc = new ethers.Contract(quoteToken, TOKENABI, readProvider);
        const dec = Number(await bc.decimals?.().then((x) => Number(x)).catch(() => 18));
        const bal = await bc.balanceOf(addr);
        setQuoteBalanceDisplay(ethers.formatUnits(bal, dec));
      } else {
        setQuoteBalanceDisplay("-");
      }
    } catch (e) {
      console.warn("failed reading base balance", e);
      setBaseBalanceDisplay("-");
      setQuoteBalanceDisplay("-");
    }
  }, [baseToken, quoteToken]);



useEffect(() => {
  if (!baseToken || !quoteToken || availableTokens.length === 0) return;
  
  setIsSwapped(false);
  
  const tokenIn = availableTokens.find(t => t.address === baseToken)?.symbol;
  const tokenOut = availableTokens.find(t => t.address === quoteToken)?.symbol;
  
  if (!tokenIn || !tokenOut) return;
  
  if (wsOrderBookRef.current?.readyState === WebSocket.OPEN) {
    console.log("WebSocket already open, skipping...");
    return;
  }
  
  const wsURL = `ws://${backendUrl}/orderbook/ws/${tokenIn}/${tokenOut}`;
  const ws = new WebSocket(wsURL);
  wsOrderBookRef.current = ws;
  
  ws.onopen = async () => {
    console.log("Connected to orderbook:", tokenIn, "/", tokenOut);
    try {
      const normalized = await getOrderBookNormalized(tokenIn, tokenOut);
      setOrderBook(normalized as any);
    } catch (err) {
      console.error("Failed to fetch orderbook:", err);
    }
  };
  
  ws.onmessage = (event) => {
    try {
      const msg = JSON.parse(event.data);
      let lastPriceToSet = "-"; // default if invalid
      if (msg.lastPrice && Number(msg.lastPrice) > 0) {
        const baseSymbol = availableTokens.find(t => t.address === baseToken)?.symbol;
        // Check if we need to invert based on whether symbols match AND isSwapped state
        if (msg.symbolIn === baseSymbol && !isSwapped) {
          lastPriceToSet = msg.lastPrice;
        } else if (msg.symbolIn !== baseSymbol && isSwapped) {
          lastPriceToSet = msg.lastPrice;
        } else {
          lastPriceToSet = (1 / Number(msg.lastPrice)).toFixed(2); // invert price for reversed pair
        }
      }
      setLastPrice(lastPriceToSet);
      const normalized = normalizeOrderBook(msg.data, tokenIn);
      setOrderBook(normalized as any);
    } catch (err) {
      console.error("WS parse error", err);
    }
  };
  
  ws.onerror = (err) => { 
    console.error("WS error", err); 
    ws.close(); 
  };
  
  ws.onclose = () => {
    console.log("Disconnected from orderbook:", tokenIn, "/", tokenOut);
  };
  
  return () => {
    console.log("Cleaning up WebSocket...");
    if (ws.readyState === WebSocket.OPEN || ws.readyState === WebSocket.CONNECTING) {
      ws.close();
    }
    wsOrderBookRef.current = null;
  };
}, [baseToken, quoteToken, isSwapped]);

  // Load token list from TokenRegistry (if VITE_TOKENREGISTRY is present) or fall back to demo list
useEffect(() => {
  // Load tokens from on-chain registry (if available) and merge with any custom tokens stored in localStorage
  const fetchAndLoad = async () => {
    try {
      let tokens: Token[] = [];

      if (registryAddr) {
        try {
          const registry = new ethers.Contract(registryAddr, REGISTRYABI, readProvider);
          const addrs: string[] = await registry.getAllTokens();

          const registryTokens = await Promise.all(
            addrs.map(async (addr: string) => {
              try {
                const c = new ethers.Contract(addr, TOKENABI, readProvider);
                const symbol = await c.symbol();
                const name = await c.name();
                return { name: name, symbol: symbol, address: addr };
              } catch (err) {
                console.log(err);
                return { name: addr, symbol: addr, address: addr };
              }
            })
          );

          tokens = registryTokens;
        } catch (err) {
          console.error("Failed to load TokenRegistry", err);
        }
      }


      if (tokens.length > 0) {
        setAvailableTokens(tokens);
        const firstValid = tokens.find((t) => t.address && t.address.startsWith("0x"))?.address ?? tokens[0].address;
        const secondValid = tokens.find((t) => t.address && t.address.startsWith("0x") && t.address !== firstValid)?.address ?? tokens[1]?.address ?? tokens[0].address;
        setBaseToken((prev) => (prev ? prev : firstValid));
        setQuoteToken((prev) => (prev ? prev : secondValid));
      }
    } catch (err) {
      console.error("Failed to initialize token list", err);
    }
  };

  fetchAndLoad();
}, []);

// Handler to add a custom token from the modal and persist it
function handleAddCustomToken(token: Token) {
  setAvailableTokens((cur) => {
    const exists = cur.some((t) => t.address.toLowerCase() === token.address.toLowerCase() || t.symbol === token.symbol);
    if (exists) return cur;
    const next = [...cur, token];
    try {
      // persist only the custom tokens (filter those with non-empty address that are likely custom)
      const storedRaw = localStorage.getItem("customTokens");
      const stored: Token[] = storedRaw ? JSON.parse(storedRaw) : [];
      stored.push(token);
      localStorage.setItem("customTokens", JSON.stringify(stored));
    } catch (e) {
      console.warn("Failed to persist custom token", e);
    }
    // if no base/quote selected, pick this as base
    if (!baseToken) setBaseToken(token.address);
    if (!quoteToken) setQuoteToken(token.address);
    return next;
  });
}

  // Trade banner state
  const [lastTrade, setLastTrade] = useState<null | {
    price: string; amountBase: string; amountQuote: string;
  }>(null);


  // Auto-hide the banner after 3 seconds whenever it appears
  useEffect(() => {
    if (!lastTrade) return;
    const t = setTimeout(() => setLastTrade(null), 3000);
    return () => clearTimeout(t);
  }, [lastTrade]);

  // Disconnect handler - clears app wallet state and closes WS
  const onDisconnect = useCallback(() => {
    // close websocket if open
    if (wsRef.current) {
      try {
        wsRef.current.close();
      } catch (e) {
        // ignore
      }
      wsRef.current = null;
    }

    setAcct("(not connected)");
    setPastOrders([]);
    setUserOrders([]);
    setBaseBalanceDisplay("-");
    setQuoteBalanceDisplay("-");
  }, []);

  // React to injected wallet account/chain changes
useEffect(() => {
  const eth = (window as any).ethereum;
  if (!eth || !eth.on) return;

  const handleAccountsChanged = async (accounts: string[] | string) => {
    const primary = Array.isArray(accounts) ? accounts[0] : accounts;

    if (!primary) {
      onDisconnect();
      return;
    }

    // Close existing WebSocket connection first
    if (wsNotifyRef.current) {
      try {
        wsNotifyRef.current.close();
      } catch (e) {
        console.error("Error closing notification WS", e);
      }
      wsNotifyRef.current = null;
    }

    setAcct(primary);
    setUserOrders([]);
    
    const orders = await getOrdersByAddress(primary);
    setUserOrders(
      (orders || []).slice().sort((a, b) => {
        const na = BigInt(a.nonce);
        const nb = BigInt(b.nonce);
        if (na > nb) return -1; // descending
        if (na < nb) return 1;
        return 0;
      })
    );

    setPastOrders([]);
    const pastOrders = await getPastHistory(primary);
    const flatOrders: Order[] = Object.values(pastOrders).flat() as Order[];
    setPastOrders(flatOrders);

    // Create new notification WebSocket
    const backendUrl = import.meta.env.VITE_BACKEND_URL;
    const wsNotify = new WebSocket(`ws://${backendUrl}/ws`);
    wsNotifyRef.current = wsNotify;

    wsNotify.onopen = () => {
      console.log("Notification WS connected for account:", primary);
      wsNotify.send(JSON.stringify({ address : primary }));
    };

    wsNotify.onmessage = (event) => {
      try {
        const msg = JSON.parse(event.data);
        console.log("Notification WS message:", msg);
        if (!(msg.event)) return;
        if (msg.event == "OrderAdd") {
          setUserOrders((cur) => {
            const exists = cur?.some(
              (o) => o.nonce === msg.data.Nonce && o.signature === msg.data.Signature
            );
            if (exists) return cur;
            // Add the new order at the top 
            if (msg.data.ConditionalOrder != "") {
              const cOrder = JSON.parse(msg.data.ConditionalOrder)
              const extractedOrder = {
                createdBy: cOrder.CreatedBy,
                symbolIn: cOrder.SymbolIn,
                symbolOut: cOrder.SymbolOut,
                amtIn: cOrder.AmtIn.toString(),
                amtOut: cOrder.AmtOut.toString(),
                nonce: cOrder.Nonce,
                status: cOrder.Status,
                filledAmtIn: cOrder.FilledAmtIn.toString(),
                limitPrice: cOrder.LimitPrice,
                signature: cOrder.Signature,
              };
              const newUserOrder: Order = {
                createdBy: msg.data.CreatedBy,
                symbolIn: msg.data.SymbolIn,
                symbolOut: msg.data.SymbolOut,
                amtIn: msg.data.AmtIn.toString(),
                amtOut: msg.data.AmtOut.toString(),
                nonce: msg.data.Nonce,
                status: msg.data.Status,
                filledAmtIn: msg.data.FilledAmtIn.toString(),
                limitPrice: msg.data.LimitPrice,
                signature: msg.data.Signature,
                conditionalOrder: extractedOrder
              };
              return [newUserOrder, ...(cur || [])];
            }
            const newUserOrder: Order = {
              createdBy: msg.data.CreatedBy,
              symbolIn: msg.data.SymbolIn,
              symbolOut: msg.data.SymbolOut,
              amtIn: msg.data.AmtIn.toString(),
              amtOut: msg.data.AmtOut.toString(),
              nonce: msg.data.Nonce,
              status: msg.data.Status,
              filledAmtIn: msg.data.FilledAmtIn.toString(),
              limitPrice: msg.data.LimitPrice,
              signature: msg.data.Signature,
            };
            return [newUserOrder, ...(cur || [])];
          });
        } else if (msg.event == "OrderRemove") {
          const nonce = BigInt(msg.data.nonce);

          setUserOrders((prev) => {
            const targetOrder = prev.find((order) => BigInt(order.nonce) === nonce);
            if (!targetOrder) return prev; // nothing to remove

            setPastOrders((past) => {
              const alreadyExists = past?.some(
                (order) =>
                  BigInt(order.nonce) === nonce &&
                  (order.status === "Completed" || order.status === "3" ||
                    order.status === "Cancelled" || order.status === "4")
              );

              if (alreadyExists) return past;

              return [
                ...(past || []),
                {
                  ...targetOrder,
                  status: "Cancelled",
                },
              ];
            });

            // Remove from active orders
            return prev.filter((order) => BigInt(order.nonce) !== nonce);
          });
        }
        else if (msg.event == "TransactionChange") {
          const nonce = BigInt(msg.data.Nonce);
          const newStatus = msg.data.Status;
          
          setUserOrders((prev) => {
            const targetOrder = prev.find((order) => BigInt(order.nonce) === nonce);
            if (!targetOrder) return prev;

            const completedOrCancelled =
              newStatus === 3 ||
              newStatus === 4 ||
              newStatus === "3" ||
              newStatus === "4" ||
              newStatus === "Completed" ||
              newStatus === "Cancelled";

            if (completedOrCancelled) {
              // Use functional update and ensure no duplicates in pastOrders
              setPastOrders((past) => {
                // Only add if not already in pastOrders
                const alreadyExists = past?.some(
                  (order) => BigInt(order.nonce) === nonce
                );
                if (alreadyExists) return past;
                fetchBalances(primary);
                return [
                  ...(past || []),
                  { ...targetOrder, status: newStatus, filledAmtIn: msg.data.FilledAmtIn },
                ];
              });

              // Remove from active orders
              return prev.filter((order) => BigInt(order.nonce) !== nonce);
            }

            // Otherwise, just update status in place
            return prev.map((order) =>
              BigInt(order.nonce) === nonce
                ? { ...order, status: newStatus }
                : order
            );
          });
        }
      } catch (e) {
        console.error("Failed to parse notification WS message", e);
      }
    };

    wsNotify.onclose = () => {
      console.log("Notification WS closed for account:", primary);
    };

    wsNotify.onerror = (e) => {
      console.error("Notification WS error", e);
    };
  };

  eth.on("accountsChanged", handleAccountsChanged);
  return () => {
    try { eth.removeListener("accountsChanged", handleAccountsChanged); } catch {}
  };
}, [onDisconnect]);

  // Connect wallet handler
async function onConnect() {
  try {
    const { address, chainId } = await connectWallet();
    setAcct(address);
    setChainId(chainId);
    
    const pastOrders = await getPastHistory(address);
    const flatOrders: Order[] = Object.values(pastOrders).flat() as Order[];
    setPastOrders(flatOrders);

    // Close and recreate notification WebSocket
    if (wsNotifyRef.current) {
      try {
        wsNotifyRef.current.close();
      } catch (e) {
        console.error("Error closing notification WS", e);
      }
      wsNotifyRef.current = null;
    }

    // Create new WebSocket immediately
    const backendUrl = import.meta.env.VITE_BACKEND_URL;
    const wsNotify = new WebSocket(`ws://${backendUrl}/ws`);
    wsNotifyRef.current = wsNotify;

    wsNotify.onopen = () => {
      console.log("Notification WS connected for account:", address);
      wsNotify.send(JSON.stringify({ address: address }));
    };

    wsNotify.onmessage = (event) => {
      try {
        const msg = JSON.parse(event.data);
        console.log("Notification WS message:", msg);
        if (!(msg.event)) return;
        if (msg.event == "OrderAdd") {
          setUserOrders((cur) => {
            const exists = cur?.some(
              (o) => o.nonce === msg.data.Nonce && o.signature === msg.data.Signature
            );
            if (exists) return cur;
            // Add the new order at the top 
            if (msg.data.ConditionalOrder != "") {
              const cOrder = JSON.parse(msg.data.ConditionalOrder)
              const extractedOrder = {
                createdBy: cOrder.CreatedBy,
                symbolIn: cOrder.SymbolIn,
                symbolOut: cOrder.SymbolOut,
                amtIn: cOrder.AmtIn.toString(),
                amtOut: cOrder.AmtOut.toString(),
                nonce: cOrder.Nonce,
                status: cOrder.Status,
                filledAmtIn: cOrder.FilledAmtIn.toString(),
                limitPrice: cOrder.LimitPrice,
                signature: cOrder.Signature,
              };
              const newUserOrder: Order = {
                createdBy: msg.data.CreatedBy,
                symbolIn: msg.data.SymbolIn,
                symbolOut: msg.data.SymbolOut,
                amtIn: msg.data.AmtIn.toString(),
                amtOut: msg.data.AmtOut.toString(),
                nonce: msg.data.Nonce,
                status: msg.data.Status,
                filledAmtIn: msg.data.FilledAmtIn.toString(),
                limitPrice: msg.data.LimitPrice,
                signature: msg.data.Signature,
                conditionalOrder: extractedOrder
              };
              return [newUserOrder, ...(cur || [])];
            }
            const newUserOrder: Order = {
              createdBy: msg.data.CreatedBy,
              symbolIn: msg.data.SymbolIn,
              symbolOut: msg.data.SymbolOut,
              amtIn: msg.data.AmtIn.toString(),
              amtOut: msg.data.AmtOut.toString(),
              nonce: msg.data.Nonce,
              status: msg.data.Status,
              filledAmtIn: msg.data.FilledAmtIn.toString(),
              limitPrice: msg.data.LimitPrice,
              signature: msg.data.Signature,
            };
            return [newUserOrder, ...(cur || [])];
          });
        } else if (msg.event == "OrderRemove") {
          const nonce = BigInt(msg.data.nonce);

          setUserOrders((prev) => {
            const targetOrder = prev.find((order) => BigInt(order.nonce) === nonce);
            if (!targetOrder) return prev; // nothing to remove

            setPastOrders((past) => {
              const alreadyExists = past?.some(
                (order) =>
                  BigInt(order.nonce) === nonce &&
                  (order.status === "Completed" || order.status === "3" ||
                    order.status === "Cancelled" || order.status === "4")
              );

              if (alreadyExists) return past;

              return [
                ...(past || []),
                {
                  ...targetOrder,
                  status: "Cancelled",
                },
              ];
            });

            // Remove from active orders
            return prev.filter((order) => BigInt(order.nonce) !== nonce);
          });
        }
        else if (msg.event == "TransactionChange") {
          const nonce = BigInt(msg.data.Nonce);
          const newStatus = msg.data.Status;

          setUserOrders((prev) => {
            const targetOrder = prev.find((order) => BigInt(order.nonce) === nonce);
            if (!targetOrder) return prev;

            const completedOrCancelled =
              newStatus === 3 ||
              newStatus === 4 ||
              newStatus === "3" ||
              newStatus === "4" ||
              newStatus === "Completed" ||
              newStatus === "Cancelled";

            if (completedOrCancelled) {
              // Use functional update and ensure no duplicates in pastOrders
              setPastOrders((past) => {
                // Only add if not already in pastOrders
                const alreadyExists = past?.some(
                  (order) => BigInt(order.nonce) === nonce
                );
                if (alreadyExists) return past;

                fetchBalances(address);
                return [
                  ...(past || []),
                  { ...targetOrder, status: newStatus, filledAmtIn: msg.data.FilledAmtIn },
                ];
              });

              // Remove from active orders
              return prev.filter((order) => BigInt(order.nonce) !== nonce);
            }

            // Otherwise, just update status in place
            return prev.map((order) =>
              BigInt(order.nonce) === nonce
                ? { ...order, status: newStatus }
                : order
            );
          });
        }
      } catch (e) {
        console.error("Failed to parse notification WS message", e);
      }
    };

    wsNotify.onclose = () => console.log("Notification WS closed");
    wsNotify.onerror = (e) => console.error("Notification WS error", e);

    if (wsRef.current) {
      try { wsRef.current.close(); } catch {}
      wsRef.current = null;
    }
    wsRef.current = openOrderWs(address, (msg: any) => handleWsMessage(msg, address));

    // Load user orders
    try {
      const orders = await getOrdersByAddress(address);
      setUserOrders(
        (orders || []).slice().sort((a, b) => {
          const na = BigInt(a.nonce);
          const nb = BigInt(b.nonce);
          if (na > nb) return -1; // descending
          if (na < nb) return 1;
          return 0;
        })
      );
    } catch (e) {
      console.warn("failed to fetch user orders", e);
    }

  } catch (e: any) {
    alert(e?.message || e);
  }
}

function handleWsMessage(msg: any, address: string) {
  try {
    if (!msg || typeof msg !== "object") return;

    if (msg.event === "TransactionChange") {
      const d = msg.data || {};
      const createdBy = d["CreatedBy"] || d["createdBy"] || address;
      const nonce = d["Nonce"] || d["nonce"] || "";
      const txHashesRaw = d["TransactionHashes"] || d["transactionHashes"] || "";
      const txs = String(txHashesRaw || "").split(",").filter(Boolean);
      const tx = txs[0];

      const filled = d["FilledAmtIn"] || d["filledAmtIn"] || d["AmtIn"] || d["amtIn"];
      const amtIn = d["AmtIn"] || d["amtIn"];
      const amtOut = d["AmtOut"] || d["amtOut"];

      let amountBase = "0";
      let amountQuote = "0";
      let price = "0";

      try {
        if (filled && amtIn && amtOut) {
          amountBase = ethers.formatUnits(filled, 18);
          const baseTotal = parseFloat(ethers.formatUnits(amtIn, 18));
          const quoteTotal = parseFloat(ethers.formatUnits(amtOut, 18));
          if (Number(amountBase) > 0) {
            price = String(quoteTotal / baseTotal);
            amountQuote = (parseFloat(price) * parseFloat(amountBase)).toFixed(6);
          }
        }
      } catch (e) {
        console.warn("failed to parse amounts", e);
      }

      const now = new Date().toLocaleTimeString("en-GB", { hour12: false, hour: "2-digit", minute: "2-digit", second: "2-digit" });
      const newTrade = {
        id: nonce || Math.random().toString(36).slice(2, 8),
        time: now,
        taker: createdBy,
        maker: "(match)",
        amountBase,
        amountQuote,
        price: Number(price).toFixed(6),
      };

      setLastTrade({ price: newTrade.price, amountBase: newTrade.amountBase, amountQuote: newTrade.amountQuote });

      // Update orders in state
      setOrders(prev => prev.map(o => o.nonce.toString() === String(nonce) ? { ...o, status: d["Status"] ?? o["status"], filled } : o));
      setUserOrders(prev => prev.map(uo => (uo["Nonce"] || uo["nonce"]) === String(nonce) ? { ...uo, Status: d["Status"] ?? uo["Status"], FilledAmtIn: d["FilledAmtIn"] ?? uo["FilledAmtIn"] } : uo));
    }

  } catch (e) {
    console.error("ws handler error", e);
  }
}

  // Refresh balances when account or selected pair changes
  useEffect(() => {
    if (!acct || acct === "(not connected)") return;
    try {
      fetchBalances(acct);
    } catch (e) {
      console.warn("failed to fetch balances on change", e);
    }
  }, [acct, baseToken, quoteToken, fetchBalances]);

  // handle real order submission: sign -> POST -> optimistic add
async function handleOrderSubmit(details: TokenQuantity) {
  try {
    // Switch to correct network first
    await switchOrAddNetwork({
      chainId: (31337).toString(16),
      chainName: "Hardhat Local",
      nativeCurrency: {
        name: "SmashCoin",
        symbol: "SC",
        decimals: 18,
      },
      rpcUrls: ["http://127.0.0.1:8545"],
      blockExplorerUrls: [""],
    });

    await (window as any).ethereum.request({ method: "eth_requestAccounts" });
    const provider = new ethers.BrowserProvider((window as any).ethereum);
    const signer = await provider.getSigner();
    const acct = await signer.getAddress();

    const network = await provider.getNetwork();
    const currentChainId = Number(network.chainId);

    const tokenInInfo = availableTokens.find(t => t.address === baseToken);
    const tokenOutInfo = availableTokens.find(t => t.address === quoteToken);
    if (!tokenInInfo || !tokenOutInfo) {
      return alert("Invalid token pair selected");
    }

    const tokenOutContract = new ethers.Contract(tokenOutInfo.address, TOKENABI, signer);

    const balance = await tokenOutContract.balanceOf(acct);
    if (balance < details.tokenOut) {
      alert("Balance is less than token out!");
      if (details.hasCondition && balance < details.conditionalTokenOut!) {
        alert("Balance is less than conditional token out!");
      }
      return;
    }

    const allowance = await tokenOutContract.allowance(acct, EXCHANGE_ADDRESS);
    if (allowance < details.tokenOut) {
      const confirmApproval = confirm(
        `You need to approve ${tokenOutInfo.symbol} first. Approve max?`
      );
      if (!confirmApproval) return;

      const tx = await tokenOutContract.approve(EXCHANGE_ADDRESS, ethers.MaxUint256);
      await tx.wait();
      alert(`${tokenOutInfo.symbol} approved successfully!`);
    }

    // === Parent order ===
    const parentNonce = BigInt(await getNextNonce(acct));

    const order: Order = {
      createdBy: acct,
      symbolIn: tokenInInfo.symbol,
      symbolOut: tokenOutInfo.symbol,
      amtIn: details.tokenIn,
      amtOut: details.tokenOut,
      nonce: parentNonce,
    };

    // === Conditional order ===
    let conditionTriggers: { stopPrice: bigint; } | undefined;

    if (
      details.hasCondition &&
      details.conditionalTokenIn &&
      details.conditionalTokenOut &&
      details.conditionalStopPrice &&
      details.conditionalLimitPrice
    ) {
      const conditionalNonce = BigInt(await getNextNonce(acct));

      const conditionalOrder: Order = {
        createdBy: acct,
        symbolIn: tokenInInfo.symbol,
        symbolOut: tokenOutInfo.symbol,
        amtIn: details.conditionalTokenIn,
        amtOut: details.conditionalTokenOut,
        nonce: conditionalNonce,
      };

      const conditionalSignature = await signOrder(signer, currentChainId, conditionalOrder);

      order.conditionalOrder = {
        ...conditionalOrder,
        signature: conditionalSignature,
      };

      // Populate the conditionTriggers parameter for the API
      conditionTriggers = {
        stopPrice: details.conditionalStopPrice,
      };
    }

    const signature = await signOrder(signer, currentChainId, order);

    // === Submit to backend with optional conditionTriggers ===
    await submitLimitOrderToBackend(order, signature, conditionTriggers);

    alert("Order submitted successfully!");
  } catch (error: any) {
    console.error("Order submission error:", error);
    alert(`Failed to submit order: ${error.message || error}`);
  }
}




  async function handleCancelOrder(order: Order) {
    const confirmation = confirm(`Delete order with nonce ${order.nonce}?`);
    if (!confirmation) return;
    try {
      await cancelOrder(order);
      // setOrders((cur) => cur.filter((local) => String(local.id) !== String(req.nonce)));
    } catch (e: any) {
      alert(e?.message || e);
    }
  }

  // UI
  return (
    <div className="min-h-screen bg-gradient-to-b from-zinc-950 via-zinc-900 to-zinc-950 text-zinc-100 p-6">
      <AnimatePresence>
        {lastTrade && (
          <motion.div
            key="trade-banner"
            initial={{ y: -60, opacity: 0 }}
            animate={{ y: 0, opacity: 1 }}
            exit={{ y: -60, opacity: 0 }}
            transition={{ duration: 0.25 }}
            className="fixed top-4 left-1/2 -translate-x-1/2 z-50 
                      bg-zinc-900/90 border border-zinc-700 backdrop-blur 
                      text-zinc-100 px-4 py-2 rounded-xl shadow-lg"
          >
            <div className="flex items-center gap-2">
              <span className="font-semibold">New trade:</span>
              <span className="font-mono">{lastTrade.amountBase}</span>
              <span className="text-zinc-400">for</span>
              <span className="font-mono">{lastTrade.amountQuote}</span>
              <span className="text-zinc-400">at</span>
              <span className="font-mono text-emerald-400">${lastTrade.price}</span>
            </div>
          </motion.div>
        )}
      </AnimatePresence>
      <div className="max-w-6xl mx-auto space-y-6">
        <h1 className="text-4xl font-extrabold mb-4 text-center text-indigo-400 drop-shadow">
          Decentralized Exchange (Demo)
        </h1>
        {/* Wallet Connect */}
        <div className="text-center">
          {acct === "(not connected)" ? (
            <button
              onClick={onConnect}
              className="mt-4 px-6 py-2 rounded-lg bg-indigo-600 hover:bg-indigo-500 text-white font-semibold shadow"
            >
              Connect Wallet
            </button>
          ) : (
            <div className="flex items-center justify-center gap-3">
              <button
                onClick={onDisconnect}
                className="mt-4 px-4 py-2 rounded-lg bg-rose-600 hover:bg-rose-500 text-white font-semibold shadow"
              >
                Disconnect
              </button>
            </div>
          )}
          <p className="text-sm text-zinc-400 mt-2">
            Account: {acct}
          </p>
        </div>
        {/* Last Price */}
        <div className="text-center mb-2">
          {(() => {
            const baseInfo = availableTokens.find(t => t.address === baseToken || t.symbol === baseToken);
            const quoteInfo = availableTokens.find(t => t.address === quoteToken || t.symbol === quoteToken);
            const baseSymbol = baseInfo?.symbol || baseToken;
            const quoteSymbol = quoteInfo?.symbol || quoteToken;
            return (
              <>
                <span className="text-sm text-zinc-400">
                  Last Price ({baseSymbol}/{quoteSymbol}): 
                </span>
                <span className="font-mono text-indigo-400 ml-2">{lastPrice}</span>
              </>
            );
          })()}
        </div>
        {/* Pair selector */}
        <div className="max-w-sm mx-auto">
          <TokenSelector
            tokens={availableTokens}
            baseToken={baseToken}
            quoteToken={quoteToken}
            setBaseToken={setBaseToken}
            setQuoteToken={setQuoteToken}
            onSwap={() => {
              // Toggle the swap state so the price calculation knows to invert
              setIsSwapped(prev => !prev);
            }}
          />
          <div className="flex justify-end mt-2">
            <button
              onClick={() => setIssueModalOpen(true)}
              className="text-xs ml-2 px-3 py-1 rounded bg-emerald-700 hover:bg-emerald-600"
            >
              + Issue token
            </button>
          </div>
          <div className="mt-3 text-sm text-zinc-300">
            <div className="flex justify-between gap-4">
              <div>
                <div className="text-xs text-zinc-400">Token balance (Base) ({availableTokens.find(t=>t.address===baseToken || t.symbol===baseToken)?.symbol || baseToken})</div>
                  <div className="font-mono">{baseBalanceDisplay}</div>
              </div>
              <div>
                <div className="text-xs text-zinc-400">Token balance (Quote) ({availableTokens.find(t=>t.address===quoteToken || t.symbol===quoteToken)?.symbol || quoteToken})</div>
                  <div className="font-mono">{quoteBalanceDisplay}</div>
              </div>
            </div>
          </div>
        </div>

    {/* Order Form */}
  <OrderForm onSubmit={handleOrderSubmit} />

    {/* Token mapping modal */}
    <IssueTokenModal isOpen={issueModalOpen} onClose={() => setIssueModalOpen(false)} onDeployed={(t) => handleAddCustomToken({ address: t.address, symbol: t.symbol, name: t.name })} />

        {/* My Orders (from backend) */}
        <div className="bg-zinc-900 rounded-2xl p-4 shadow-lg border border-zinc-800 mt-6">
          <h3 className="font-semibold text-indigo-400 mb-2">My Orders</h3>
          {userOrders.length === 0 ? (
            <div className="text-sm text-zinc-400">You have no active orders.</div>
          ) : (
            // limit height and allow scrolling for long lists
            <div className="space-y-2 max-h-72 overflow-auto pr-2">
            {userOrders.map((order) => (
              <div
                key={order.nonce || Math.random().toString(36).slice(2, 8)}
                className="bg-zinc-800/40 p-3 rounded-lg flex items-start gap-4"
              >
                {/* Left section: main + conditional orders */}
                <div className="flex gap-4 flex-1">
                  {/* Parent Order */}
                  <div className="text-sm space-y-1">
                    <div className="font-mono text-xs text-zinc-300">Nonce: {order.nonce}</div>
                    <div className="text-xs">Pair: {order.symbolIn}/{order.symbolOut}</div>
                    <div className="text-xs">In: {ethers.formatUnits(order.amtIn || "0", 18)}</div>
                    <div className="text-xs">Out: {ethers.formatUnits(order.amtOut || "0", 18)}</div>
                    <div className="text-xs">Filled Amt: {ethers.formatUnits(order.filledAmtIn || "0", 18)}</div>
                    <div className="text-xs text-zinc-400">Status: {order.status || "Unknown"}</div>
                  </div>

                  {/* Conditional Order */}
                  {order.conditionalOrder && (
                    <div className="border-l-2 border-indigo-500 pl-3 text-xs text-zinc-400 space-y-1">
                      <div className="font-semibold text-indigo-400 mb-1">Conditional Order</div>
                      <div>Nonce: {order.conditionalOrder.nonce}</div>
                      <div>Pair: {order.conditionalOrder.symbolIn}/{order.conditionalOrder.symbolOut}</div>
                      <div>In: {ethers.formatUnits(order.conditionalOrder.amtIn || "0", 18)}</div>
                      <div>Out: {ethers.formatUnits(order.conditionalOrder.amtOut || "0", 18)}</div>
                    </div>
                  )}
                </div>

                {/* Cancel Button */}
                <div className="flex-shrink-0">
                  {(order.status !== "Completed" && order.status !== "Cancelled") && (
                    <button
                      onClick={() => handleCancelOrder(order)}
                      className="px-3 py-1 rounded bg-rose-600 hover:bg-rose-500 text-sm"
                    >
                      Cancel
                    </button>
                  )}
                </div>
              </div>
            ))}
            </div>
          )}
        </div>


        {/* DEX Panels */}
        <div className="grid md:grid-cols-2 gap-6 mt-8">
          <OrderBook orderbooks={orderBook} />
          <TradeHistory trades={pastOrders} />
        </div>

        <p className="text-zinc-500 text-xs text-center mt-6">
          RPC: {import.meta.env.VITE_RPC_URL}
        </p>
      </div>
    </div>
  );
}