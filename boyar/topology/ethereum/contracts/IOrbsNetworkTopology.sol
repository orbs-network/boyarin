pragma solidity 0.5.3;

interface IOrbsNetworkTopology {
    function getNetworkTopology() external returns (address[] memory, bytes4[] memory);
}