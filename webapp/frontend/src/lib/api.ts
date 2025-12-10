import { type Signer } from "ethers";
import type { Order } from "../App";

const BACKEND = "http://" + (import.meta as any).env.VITE_BACKEND_URL || "http://localhost:8080";
const EXCHANGE_ADDRESS = (import.meta as any).env.VITE_EXCHANGE || "";

export async function getNextNonce(address: string) {
  const res = await fetch(`${BACKEND}/nonce`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ address }),
  });
  if (!res.ok) throw new Error("Failed to get nonce from backend");
  const body = await res.json();
  return String(body.nonce);
}

export async function signOrder(
  signer: Signer,
  chainId: number,
  orderObj: Order | {
    createdBy: string;
    symbolIn: string;
    symbolOut: string;
    amtIn: bigint;
    amtOut: bigint;
    nonce: bigint;
  }
) {
  const domain = {
    name: "SmashDEX",
    version: "1",
    chainId,
    verifyingContract: EXCHANGE_ADDRESS,
  };

  const types = {
    Order: [
      { name: "createdBy", type: "address" },
      { name: "symbolIn", type: "string" },
      { name: "symbolOut", type: "string" },
      { name: "amtIn", type: "uint256" },
      { name: "amtOut", type: "uint256" },
      { name: "nonce", type: "uint256" },
    ],
  };

  const value = {
    createdBy: orderObj.createdBy,          
    symbolIn: orderObj.symbolIn,            
    symbolOut: orderObj.symbolOut,          
    amtIn: orderObj.amtIn,          
    amtOut: orderObj.amtOut,
    nonce: orderObj.nonce
  };  

  const sig = await signer.signTypedData(domain, types, value);
  return sig;
}

export async function submitLimitOrderToBackend(
  order: Order, 
  signature: string,
  condition?: {
    stopPrice: bigint,
  }
) {
  const serialized_order: any = {
    createdBy: order.createdBy,
    symbolIn: order.symbolIn,
    symbolOut: order.symbolOut,
    amtIn: order.amtIn.toString(),  
    amtOut: order.amtOut.toString(),
    nonce: order.nonce.toString(),
  };
  
  if (order.conditionalOrder) {
    serialized_order.conditionalOrder = {
      order: {
        createdBy: order.conditionalOrder.createdBy,
        symbolIn: order.conditionalOrder.symbolIn,
        symbolOut: order.conditionalOrder.symbolOut,
        amtIn: order.conditionalOrder.amtIn.toString(),
        amtOut: order.conditionalOrder.amtOut.toString(),
        nonce: order.conditionalOrder.nonce.toString(),
      },
      signature: order.conditionalOrder.signature,
    };
  }
  
  const requestBody = { 
    order: serialized_order, 
    signature: signature,
    conditionTriggers: {},
  };

  if (condition) {
    requestBody.conditionTriggers = {stopPrice: condition.stopPrice.toString()};
  }
  
  const res = await fetch(`${BACKEND}/order/limit`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(requestBody),
  });
  
  if (!res.ok) {
    const txt = await res.text();
    throw new Error(`Send order failed: ${res.status} ${txt}`);
  }
  return true;
}

function keysToCamel<T>(obj: any): T {
  if (Array.isArray(obj)) {
    return obj.map((v) => keysToCamel(v)) as any;
  } else if (obj !== null && obj.constructor === Object) {
    return Object.fromEntries(
      Object.entries(obj).map(([k, v]) => [
        k.charAt(0).toLowerCase() + k.slice(1),
        keysToCamel(v),
      ])
    ) as T;
  }
  return obj;
}

export async function getOrdersByAddress(address: string) {
  const res = await fetch(`${BACKEND}/order/${address}`, {
    method: "GET",
    headers: { "Content-Type": "application/json" },
  });

  if (!res.ok) {
    const txt = await res.text();
    throw new Error(`Get orders failed: ${res.status} ${txt}`);
  }

  const body = await res.json();

  for (const order of body.orders) {
    if (order.ConditionalOrder && order.ConditionalOrder !== "") {
      try {
        const cOrder = JSON.parse(order.ConditionalOrder);
        order.conditionalOrder = {
          createdBy: cOrder.CreatedBy,
          symbolIn: cOrder.SymbolIn,
          symbolOut: cOrder.SymbolOut,
          amtIn: cOrder.AmtIn?.toString(),
          amtOut: cOrder.AmtOut?.toString(),
          nonce: cOrder.Nonce,
          status: cOrder.Status,
          filledAmtIn: cOrder.FilledAmtIn?.toString(),
          limitPrice: cOrder.LimitPrice,
          signature: cOrder.Signature,
        };
      } catch (err) {
        console.error("Failed to parse conditionalOrder:", order.ConditionalOrder, err);
        order.ConditionalOrder = null;
      }
    } else {
      order.ConditionalOrder = null;
    }
  }

  return keysToCamel<Array<Order>>(body.orders);
}


export async function cancelOrder(orderReq: Order) {
  const values = {
    createdBy: orderReq.createdBy,
    nonce: orderReq.nonce,
    limitPrice: orderReq.limitPrice,
    symbolIn: orderReq.symbolIn,
    symbolOut: orderReq.symbolOut
  }
  const res = await fetch(`${BACKEND}/order`, {
    method: "DELETE",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(values),
  });
  if (!res.ok) {
    const txt = await res.text();
    throw new Error(`Cancel order failed: ${res.status} ${txt}`);
  }
  return true;
}

export async function getOrderBook(symbolIn: string, symbolOut: string) {
  const res = await fetch(`${BACKEND}/order/${symbolIn}/${symbolOut}`, {
    method: "GET",
    headers: { "Content-Type": "application/json" },
  });
  if (!res.ok) {
    const txt = await res.text();
    throw new Error(`Cancel order failed: ${res.status} ${txt}`);
  }
  return await res.json();
}

// Normalize backend orderbook shape to frontend OrderBooks shape
export function normalizeOrderBook(raw: any, symbolIn: string) {
  const backendBaseSymbol = raw.symbolIn;   // base
  const backendQuoteSymbol = raw.symbolOut; // quote

  // Bids: users buying base (paying quote)
  const backendBids = (raw?.bids || []).map((b: any, i: number) => ({
    id: `bid-${b.price}-${i}`,
    side: "buy" as const,
    price: Number(b.price),        // quote per base
    quantity: Number(b.quantity),  // base amount
    orderCount: Number(b.orders) || Number(b.orderCount) || 0,
  }));

  // Asks: users selling base (receiving quote)
  const backendAsks = (raw?.asks || []).map((a: any, i: number) => ({
    id: `ask-${a.price}-${i}`,
    side: "sell" as const,
    price: Number(a.price),
    quantity: Number(a.quantity),
    orderCount: Number(a.orders) || Number(a.orderCount) || 0,
  }));

  // If user-view base != backend base, we invert
  const isInverted = symbolIn !== backendBaseSymbol;

  if (!isInverted) {
    // Normal orientation (same as backend)
    return {
      baseSymbol: backendBaseSymbol,
      quoteSymbol: backendQuoteSymbol,
      bid: backendBids,
      ask: backendAsks,
    };
  }

  // Inverted view (user flipped base/quote)
  const invertedBids = backendAsks.map((a, i) => ({
    ...a,
    id: `inv-bid-${a.price}-${i}`,
    side: "buy" as const,
    // invert price (now base per quote)
    price: a.price === 0 ? 0 : 1 / a.price,
  }));

  const invertedAsks = backendBids.map((b, i) => ({
    ...b,
    id: `inv-ask-${b.price}-${i}`,
    side: "sell" as const,
    price: b.price === 0 ? 0 : 1 / b.price,
  }));

  return {
    baseSymbol: backendQuoteSymbol,
    quoteSymbol: backendBaseSymbol,
    bid: invertedBids,
    ask: invertedAsks,
  };
}

export async function getOrderBookNormalized(symbolIn: string, symbolOut: string) {
  const raw = await getOrderBook(symbolIn, symbolOut);
  return normalizeOrderBook(raw, symbolIn);
}

export function openOrderWs(address: string, onMessage: (msg: unknown) => void) {
  const url = new URL(BACKEND);
  const wsproto = url.protocol === "https:" ? "wss:" : "ws:";
  const host = url.host;
  const path = url.pathname.replace(/\/$/, "");
  const ws = new WebSocket(`${wsproto}//${host}${path}/ws`);

  ws.onopen = () => {
    try {
      ws.send(JSON.stringify({ address }));
    } catch (e) {
      console.warn("ws send subscribe failed", e);
    }
  };
  ws.onmessage = (ev) => {
    try {
      const msg = JSON.parse(ev.data as string);
      onMessage(msg);
    } catch (e) {
      console.error("bad ws message", e);
    }
  };
  ws.onclose = () => console.log("backend ws closed");
  ws.onerror = (e) => console.warn("backend ws error", e);
  return ws;
}

export async function getPastHistory(address: string) {
  const res = await fetch(`${BACKEND}/order/history/${address}`, {
    method: "GET",
    headers: { "Content-Type": "application/json" },
  });
  if (!res.ok) {
    const txt = await res.text();
    throw new Error(`Cancel order failed: ${res.status} ${txt}`);
  }
  return await res.json();
}

export { BACKEND, EXCHANGE_ADDRESS };

export async function addToken(name: string, symbol: string, address: string) {
  const req = {
    name: name,
    symbol: symbol,
    address: address
  }
  
  const res = await fetch(`${BACKEND}/token/add`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(req),
  });
  if (!res.ok) {
    const txt = await res.text();
    throw new Error(`Add token failed: ${res.status} ${txt}`);
  }
  return true;
}