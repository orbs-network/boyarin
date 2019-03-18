pragma solidity 0.5.3;

import "./IOrbsNetworkTopology.sol";


contract MockOrbsNetwork is IOrbsNetworkTopology {

bytes20[] nodes = new bytes20[](1);

bytes4[] ips = new bytes4[](1);

constructor() public  {
    nodes[0] = hex"6e2cb55e4cbe97bf5b1e731d51cc2c285d83cbf9";
    ips[0] = hex"0DEA8F0F";
}

function getNetworkTopology()
public
view
returns (bytes20[] memory nodeAddresses, bytes4[] memory ipAddresses)
{
return (nodes, ips);
}
}
