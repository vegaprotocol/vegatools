## Network History Divergence Tool

A tool to quickly identify the source of network history divergence between 2 nodes.

For example, to figure where and why the network history started to diverge between two datanodes:

`networkhistorydivergencetool https://vega.mainnet.stakingcabin.com https://m0.vega.community`

Would give output like the following:

```
Truth Server: https://vega.mainnet.stakingcabin.com
To Compare Server: https://m0.vega.community
IPFS Host: localhost:7001
Minimum Height: 0
Maximum Height: 1000000000
First Different HistorySegmentID values for FromHeight 12601 ToHeight 12900  Truth:QmbVWV9x5y9PVAiabh3ajBeEfUyGut9AEVP3rTM2gwM9wE  Compare:QmPETRkkLLGgMRsXL1UHWrXyXqx3oQWMDEwwSpPHBhphEi:
History segment QmbVWV9x5y9PVAiabh3ajBeEfUyGut9AEVP3rTM2gwM9wE sourced successfully.
History segment QmPETRkkLLGgMRsXL1UHWrXyXqx3oQWMDEwwSpPHBhphEi sourced successfully.

MISMATCHED DATA: currentstate/delegations_current at fromHeight 12601, toHeight 12900, to see differences: diff /home/matthewpendrey/projects/vegatoolstest/vegatools/networkhistorydivergencetool/segments/QmbVWV9x5y9PVAiabh3ajBeEfUyGut9AEVP3rTM2gwM9wE/currentstate/delegations_current /home/matthewpendrey/projects/vegatoolstest/vegatools/networkhistorydivergencetool/segments/QmPETRkkLLGgMRsXL1UHWrXyXqx3oQWMDEwwSpPHBhphEi/currentstate/delegations_current  

MISMATCHED DATA: history/delegations at fromHeight 12601, toHeight 12900, to see differences: diff /home/matthewpendrey/projects/vegatoolstest/vegatools/networkhistorydivergencetool/segments/QmbVWV9x5y9PVAiabh3ajBeEfUyGut9AEVP3rTM2gwM9wE/history/delegations /home/matthewpendrey/projects/vegatoolstest/vegatools/networkhistorydivergencetool/segments/QmPETRkkLLGgMRsXL1UHWrXyXqx3oQWMDEwwSpPHBhphEi/history/delegations 
```

**Prior to running the tool you must start an IPFS daemon that can connect to and source the segments from the network.** 
To do this, install ipfs: https://docs.ipfs.tech/install/command-line/#system-requirements

Update the ipfs config file (usually in ~/.ipfs) and add the relevant bootstrap peers to connect to the vega network, e.g.:

```
"Bootstrap": [
"/dns/api0.vega.community/tcp/4001/ipfs/12D3KooWAHkKJfX7rt1pAuGebP9g2BGTT5w7peFGyWd2QbpyZwaw",
    "/dns/api1.vega.community/tcp/4001/ipfs/12D3KooWDZrusS1p2XyJDbCaWkVDCk2wJaKi6tNb4bjgSHo9yi5Q",
    "/dns/api2.vega.community/tcp/4001/ipfs/12D3KooWEH9pQd6P7RgNEpwbRyavWcwrAdiy9etivXqQZzd7Jkrh",
    "/dns/m0.vega.community/tcp/4001/ipfs/12D3KooWQvja5nUKkdBR9FdX9p68B8W5ze9TeWjnuVBPPDoXcvW4",
    "/dns/m2.vega.community/tcp/4001/ipfs/12D3KooWAdkdG39tiZ7o8CoF4ktPpuxGkfG6at6YrfFwvgAwGHAv",
    "/dns/metabase.vega.community/tcp/4001/ipfs/12D3KooWCxuWEb2uGjcZpP87xprJELpZtKDXKkb2AjsYBbXZarG9",
    "/ip4/78.141.194.186/tcp/4001/p2p/12D3KooWKjggJGpS5kzQaNz7JPfGBHHQocP3qGYWhquFw8zQLZGv",
    "/dns/vega.mainnet.stakingcabin.com/tcp/4001/p2p/12D3KooWJiJoL3Ua6QYPhV7Q227r4hQvAArxLuY6NqkpMn3gFtLf",
    "/dns/vega-data.nodes.guru/tcp/4001/p2p/12D3KooWBwQLm9ZskZPDveeNMF42bskZ3M3HXPyVRJ5XtmG1todg",
    "/dns/vega-mainnet.anyvalid.com/tcp/4001/p2p/12D3KooWLxQHbaWrtW4XuzzC2DHp3qGBKhUrSscmW59FPXwtw5DW"
  ], 
```
Add the `swarm.key` file for the vega network to the ipfs home directory  (again, usually ~/.ipfs).  You can usually grab this from `<data-node state home>/networkhistory/store/ipfs/swarm.key`

Once this is done, try executing `ipfs get <segment id>` where `<segment id>` matches the id of one segment from each server, this will confirm that your IPFS setup is correct.  Once this is done you can run the tool.