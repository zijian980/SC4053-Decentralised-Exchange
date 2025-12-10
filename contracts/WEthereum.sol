// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import "@openzeppelin/contracts/token/ERC20/ERC20.sol";
import "@openzeppelin/contracts/access/Ownable.sol";
import "@openzeppelin/contracts/token/ERC20/extensions/ERC20Permit.sol";

contract WEthereum is ERC20, ERC20Permit, Ownable {
    constructor(uint256 initialSupply) ERC20("WEthereum", "WETH") Ownable(msg.sender) ERC20Permit("WEthereum") {
        _mint(msg.sender, initialSupply);
    }
}

