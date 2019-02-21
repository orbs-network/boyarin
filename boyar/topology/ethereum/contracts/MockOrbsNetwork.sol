pragma solidity 0.5.4;

import "./IOrbsNetworkTopology.sol";

contract MockOrbsNetwork is IOrbsNetworkTopology {

    address[] nodes = [address(0)];
    bytes4[] ips = [bytes4(0xFFFFFFFF)];

    function getNetworkTopology() public returns (address[] memory NodeAddresses, bytes4[] memory IpAddresses) {
        return (nodes, ips);
    }
}