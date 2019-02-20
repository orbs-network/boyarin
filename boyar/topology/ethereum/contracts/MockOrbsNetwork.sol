pragma solidity 0.5.3;

import "./IOrbsNetwork.sol";

contract MockOrbsNetwork is IOrbsNetwork {

    address[] nodes = [address(0)];
    bytes4[] ips = [bytes4(0xFFFFFFFF)];

    function getNetowrk() public returns (address[] memory, bytes4[] memory) {
        return (nodes, ips);
    }
}