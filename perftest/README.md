# vegatools perftest
This sub command is used to stress a working Vega network to allow us to capture the resource impact and throughput of the network. This command is part of the system used in the nightly run that generates performance data which we use to look for regressions.

[Grafana Performance Charts](https://monitoring.vega.community/d/a3af676b-4e80-448f-9bfa-188e46f2a800/daniel-test?orgId=1&from=now-7d&to=now)

[Nightly Performance Jenkins Pipeline](https://jenkins.vega.rocks/blue/organizations/jenkins/common%2Fperformance-tests/activity/)

## Command like options

### -a [address of the data node server]
The perf tools needs to access the data node for querying data about the network such as defined assets and markets and the state of proposals.

### -w [address of the wallet server]
The address of the vega wallet so that the tool can send transactions. The wallet has a proof of work system included to prevent transaction spam. For the performance testing we turn off or reduce to the lowest level possible the spam settings so they do not impact the results.

### -f [address of the faucet server]
To generate funds for the users to trade we use the Vega faucet. At startup the tool will create finds for every user that will be used to place trades during the test run.

### -g [address of the ganache server]
The generate Vega tokens for use in governance the tool uses a ganache server as a faucet. There needs to be enough voting users and tokens for those users for proposals to pass.

### -t [path to api token keys file]
Automated trading bots require predefined tokens that are set up when the accounts are created in the wallet. We point the tool towards a single file containing all the tokens that were created when the Vega network was started.

### -c [commands per second]
The number of transaction sent to the wallet per second. This includes new orders and cancels. The tool will attempt to spread the orders out through each second to create a more realistic scenario. If batch orders are being used, this value represents the number of batches being sent.

### -r [runtime in seconds]
The number of seconds the tool will run for once trading is ready. After this time the tool will exit.

### -n [number of normal users to send transactions with]
The number of users that are used to send orders. There must be at least this many lines in the token file for this to succeed. Normal users do not supply liquidity and do not vote on proposals.

### -u [number of lp users to provide liquidity]
The number of users that supply liquidity to the markets. Each lp user will supply the same amount of liquidity.

### -m [number of markets to create and use]
The number of individual markets to create and use. Creating a market is slow (around a minute each). Users will select a market at random to send transactions too. Each market will have the same settings and the same number of liquidity providers.

### -v [number of accounts to assign voting power]
The number of user given vote tokens that can vote on proposals. All voters will vote on each proposal and will always vote yes. The only proposals used by the tools are for new markets.

### -b [set size of batch orders]
The number of transactions in each batch. Each batch consists of a single cancel all followed by the random set of new orders. So if the size of the batch is 20, there will be 1 cancel and 19 new order transactions.

### -p [set number of pegged orders to load the market with]
The number of pegged orders per market. A random normal user will be selected to send each of the pegged orders which will be randomly configured in terms of price, size and book side.

### -L [number of price levels per side]
The number of unique price levels on each side of the book in a market. A LIMIT order per price level is placed using a random normal user. The price levels are created either side of the starting mid price and move in increments of 1. Therefore if the market mid price was 10,000 and we set the price levels to 10, we would have 20 orders on the book from 9,990 to 10009 inclusive.

### -S [number of seconds between SLA updates]
If we are using non batched orders then we will update the SLA state at this given second interval. Updating the SLA involves creating a batch of new orders submitting them, followed by a transaction to reset the liquidity commitment.

### -l [number of price levels for SLA orders]
When sending in the batch of orders to cover the liquidity commitment of the SLA, this parameter sets the number of price levels on each side of the book. 

### -d [number of stop orders per user]
This is the number of stop orders that are placed for each normal user in each market. If we have 5 markets with 100 users there will be 5*100*[number of stop order] placed into the network.

### -s [mid price to use at the start]
The starting mid price used to anchor all the other orders from. This value needs to be larger than the number of price levels or startup will fail. 

### -M [allow the mid price we place orders around to move randomly]
Allow the mid point price to move by 1 point in a random direction every time a cancel is sent. The price is limited to 500 point move in total and will be capped at this.

### -F [place an order at every available price level]
For each market, using the number of price levels parameter, place a persistent order at each level using a random normal user for each order. This is used to test how the book will handle having many price levels in use at the same time.

### -i [initialise everything and then exit]
Perform the asset generation, market creation and pre trading order steps and then exit the program. This option is to help get around timing issues where the first steps can take wildly different amount of times. This allows a user to set up the markets in the way they want and then run the tool a second time to start placing the orders.

### -I [skip the initialise steps]
Do not initialise the assets, markets or price levels but instead move to the load sending steps. THis option is usually specified after the tool has been run a first time to set everything up.

### -U [allow lp users to place orders during setup]
Treat lpusers as normal users when sending orders.

### -B [all transactions are sent in batches]
Group up transactions for each user and send them as batches each second.

### -P [use spot markets]
Create a spot markets instead of futures markets.

### -A [send AMMs]
Send AMM commitments for every market and every LP user in each market.