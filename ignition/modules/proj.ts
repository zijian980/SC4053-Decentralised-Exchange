import { buildModule } from "@nomicfoundation/hardhat-ignition/modules";

export default buildModule("TokenModule", (m) => {
  // Account #0:  0xf39fd6e51aad88f6f4ce6ab8827279cfffb92266 (10000 ETH)
  // Private Key: 0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80
  const dexDeployer = m.getAccount(0);
  // Account #1:  0x70997970c51812dc3a010c7d01b50e0d17dc79c8 (10000 ETH)
  // Private Key: 0x59c6995e998f97a5a0044966f0945389dc9e86dae88c7a8412f4603b6b78690d
  const scDeployer = m.getAccount(1);
  // Account #2:  0x3c44cdddb6a900fa2b585dd299e03d12fa4293bc (10000 ETH)
  // Private Key: 0x5de4111afa1a4b94908f83103eb1f1706367c2e68ca870fc3fb9a804cdab365a
  const dcDeployer = m.getAccount(2);
  // Account #3:  0x90f79bf6eb2c4f870365e785982e1f101e93b906 (10000 ETH)
  // Private Key: 0x7c852118294e51e653712a81e05800f419141751be58f605c371e15141b007a6
  const maker = m.getAccount(3);
  // Account #4:  0x15d34aaf54267db7d7c367839aaf71a00a2c6a65 (10000 ETH)
  // Private Key: 0x47e179ec197488593b187f80a00eb0da91f1b9d0b13f8733639f19c30a34926a  
  const taker = m.getAccount(4);
  // Account #5:  0x9965507d1a55bcc2695c58ba16fb37d819b0a4dc (10000 ETH)
  // Private Key: 0x8b3a350cf5c34c9194ca85829a2df0ec3153be0318b5e2d3348e872092edffba
  const taker2 = m.getAccount(5);
  const totalSupply = 1_000_000n * 10n ** 18n;
  const registry = m.contract("TokenRegistry", [], { from: dexDeployer } );
  const dex = m.contract("Exchange", [registry], { from: dexDeployer })
  const wEth = m.contract("WEthereum", [totalSupply], { from: dexDeployer });
  const smashCoin = m.contract("SmashCoin",[totalSupply], { from: scDeployer });
  const doogeCoin = m.contract("DoogeCoin",[totalSupply], { from: dcDeployer });
  
  m.call(registry, "addToken", [wEth], { from: dexDeployer, id: "addWEth" })
  m.call(registry, "addToken", [smashCoin], { from: dexDeployer, id: "addSmashCoin" })
  m.call(registry, "addToken", [doogeCoin], { from: dexDeployer, id: "addDoogeCoin" })

  const toSend = 5000n * 10n ** 18n
  m.call(smashCoin, "transfer", [maker, toSend], { from: scDeployer });
  m.call(doogeCoin, "transfer", [taker, toSend], { from: dcDeployer });
  m.call(wEth, "transfer", [taker2, toSend], { from: dexDeployer });
  // const toApprove = 1_000_000n * 10n ** 18n
  // m.call(smashCoin, "approve", [dex, toApprove], {from: maker});
  // m.call(doogeCoin, "approve", [dex, toApprove], {from: taker});
  // m.call(wEth, "approve", [dex, toApprove], {from: taker2});

  return { registry, dex, smashCoin, doogeCoin, wEth };
});

// On sepolia
// TokenModule#DoogeCoin - 0x7964BA64934e7aa1eC78450F2EdE4ecAAaef3EC0
//TokenModule#SmashCoin - 0x94e25ee9fb5CcBf79a47Ed1bB7Ef20dEF615d59F
//TokenModule#TokenRegistry - 0x430C94108720E0DeF477be135bf71BC40c2994BE
//TokenModule#WEthereum - 0xFaFc71f887A1eEF7a3e03945F604a2F1d7DBF8d4
//TokenModule#Exchange - 0xc808c0d65Ae46baDA65c944E5576E7936f1e7D8b
