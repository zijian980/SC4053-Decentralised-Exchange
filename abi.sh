#!/bin/bash

npx hardhat compile

jq '.abi' artifacts/contracts/SmashCoin.sol/SmashCoin.json > artifacts/contracts/SmashCoin.sol/SmashCoin.abi
abigen --abi=artifacts/contracts/SmashCoin.sol/SmashCoin.abi --pkg=token --type token --out=webapp/backend/abi/token/token.go
jq '.abi' artifacts/contracts/TokenRegistry.sol/TokenRegistry.json > artifacts/contracts/TokenRegistry.sol/TokenRegistry.abi
abigen --abi=artifacts/contracts/TokenRegistry.sol/TokenRegistry.abi --pkg=registry --type registry --out=webapp/backend/abi/registry/registry.go
jq '.abi' artifacts/contracts/Exchange.sol/Exchange.json > artifacts/contracts/Exchange.sol/Exchange.abi
abigen --abi=artifacts/contracts/Exchange.sol/Exchange.abi --pkg=exchange --type exchange --out=webapp/backend/abi/exchange/exchange.go

cp "./artifacts/contracts/Exchange.sol/Exchange.json" "./webapp/frontend/src/abi/Exchange.json"
cp "./artifacts/contracts/SmashCoin.sol/SmashCoin.json" "./webapp/frontend/src/abi/Token.json"
cp "./artifacts/contracts/TokenRegistry.sol/TokenRegistry.json" "./webapp/frontend/src/abi/TokenRegistry.json"
