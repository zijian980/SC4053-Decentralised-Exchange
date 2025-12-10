# SC4053 DEX Project

## Project Overview

Our project consists of the following components:
1. Hardhat development environment - Smart contracts development and local node
2. React frontend with Ethers.js
3. Go backend with Go-eth (GETH) and Echo web framework

We feature an orderbook-style DEX that has users submit orders, which are signed using EIP-712 standard, and verified on the backend and on on-chain before execution (1-1 match only, ring match is only executable by backend due to required logistics using DEX private key, so it is trusted).

## Features

1. Full fill
2. Partial Fill
3. Bonus (1) - Cancel order
4. Bonus (2) - Ring matching between 3 orders
5. Bonus (3) - Conditional Order (Stop-Limit)
6. Implements EIP-712 for signing data and verification on backend and blockchain
7. Create token asset (executed by EOA)

To accomplish the features regarding orders, we utilise Go's stellar concurrency capabilities to run multiple Oracles and a matching engine which match transactions 1 to 1, and if they can't, try to match them in a ring/cycle (bonus 2) with other order books. Another oracle also watches the last executed price for all orderbook, and execute conditional orders (we only have stop-limit implemented in the same direction), such that we the conditional order will be listed at the trigger price.

## Setup

To setup, go to the root of the project and do `npm i` to install dependencies for the hardhat environment. 

Then, navigate to webapp/frontend and do `npm i` to install dependencies for the frontend. 

Before starting the backend, have a terminal window open at the root of the project and do `npx hardhat node`, which will start up a local Eth node. Then, do `npx hardhat ignition deploy ./ignition/modules/proj.ts --network localhost` from another terminal, which will deploy to the local node. 

Then, have another terminal window open at webapp/backend and either run main.exe, or do `go run cmd/main.go` if you have Go installed. 

Navigate to webapp/frontend and do npm run dev to start running the frontend. 

Env files have already been provided as the deployment addresses do not change on hardhat. If deployment addresses are somehow different, you will need to replace the .env values in webapp/frontend and webapp/backend.

You will need to install Metamask and import the local node as a network after starting the node, or restart the browser if you load the node after starting up your browser and it is stuck on connecting. If the below image is not showing, the Network name can be anything, Default RPC URL http://127.0.0.1:8545, Symbol SC.

<img width="472" height="763" alt="image" src="https://github.com/user-attachments/assets/64079a62-ce52-4e5d-8609-8528db96c618" />

After setting up Metamask, import the following private keys for usage
```
0x7c852118294e51e653712a81e05800f419141751be58f605c371e15141b007a6 - Has SC
0x47e179ec197488593b187f80a00eb0da91f1b9d0b13f8733639f19c30a34926a - Has DC
0x8b3a350cf5c34c9194ca85829a2df0ec3153be0318b5e2d3348e872092edffba - Has WETH
```

If the private keys for deploying the DEX is required, it is in /ignition/modules/proj.ts, or in webapp/backend/cmd/main.go. 

Frontend URL will be `http://localhost:5173`

## Project Structure

contracts/ - contains smart contracts - Tokens, ITokenRegistry, TokenRegistry, Exchange

webapp/ - contains frontend and backend code

## Learning points
It is difficult in a DEX, especially since user is not always online to execute orders in an orderbook style DEX. Through this project, we learned why DEX has evolved to using AMM and PMM style, as well as why development is more focused on L2 than on L1. When developing on L1, there are many things to think about when doing a DEX, such as having to think about block mining times, gas fees, and having to care about orphan transactions (although rare especially in current day PoS ETH). While running in a local node using hardhat helps to alleviate these issues for our project, having to mind these issues when executing our project was a major factor. One such lesson is in doing approve before being able to use the token for future orders, which would incur gas fees (which could be heavy depending on when the user does it), and also block mining times also means that we have to wait for each transaction to be mined. Such problems affect orderbook style the most as having to ensure that all transactions are valid is key.

Using approve in itself also poses a big issue, as deciding calling per transaction (incurring more gas fees) or approving to max amount (dangerous), is one of the disadvantages of developing on L1. Another point about gas is the decision of off-chain matching vs on-chain matching, On L1, on-chain matching is inefficient and expensive, especially if we tried to implement ring matching, which while deterministic, is unpredictable. On top of all these, it is not guaranteed for a transaction to be picked up and on a mined block ASAP, further reducing the efficiency of a L1 DEX, and increasing waiting time, resulting in bad UX.


## AI Usage

Claude was used to help and debug 1-1 matching and ring matching code on the backend, as well as state management issues on the frontend
