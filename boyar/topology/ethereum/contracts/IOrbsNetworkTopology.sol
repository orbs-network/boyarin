pragma solidity 0.5.4;

interface IOrbsNetworkTopology {
    function getNetworkTopology() external returns (address[] memory, bytes4[] memory);
}