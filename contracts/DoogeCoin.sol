// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import "@openzeppelin/contracts/token/ERC20/ERC20.sol";
import "@openzeppelin/contracts/access/Ownable.sol";
import "@openzeppelin/contracts/token/ERC20/extensions/ERC20Permit.sol";

contract DoogeCoin is ERC20, ERC20Permit, Ownable {
    constructor(uint256 initialSupply) ERC20("DoogeCoin", "DC") Ownable(msg.sender) ERC20Permit("DoogeCoin") {
        _mint(msg.sender, initialSupply);
    }
}
