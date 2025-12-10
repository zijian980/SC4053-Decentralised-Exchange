// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

struct TokenDetails {
    string name;
    string symbol;
}

interface ITokenRegistry {
    function addToken(address token) external;
    function removeToken(address token) external;
    function isSupported(address token) external view returns (bool);
    function getAllTokens() external view returns (address[] memory);
    function getTokenAddress(string calldata symbol) external view returns (address);
}
