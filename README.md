# Blockchain Explorer for the Coda Protocol
The explorer provides a comprehensive and easy to use interface for the upcoming coda blockchain. It makes it easy to view blocks, follow transactions and monitor your snarking activity.

[![Badge](https://github.com/gobitfly/coda-explorer/workflows/Build/badge.svg)](https://github.com/gobitfly/coda-explorer/actions?query=workflow%3A%22Build+%26+Publish+Docker+images%22)
[![Go Report Card](https://goreportcard.com/badge/github.com/gobitfly/coda-explorer)](https://goreportcard.com/report/github.com/gobitfly/coda-explorer)

## About

The explorer is built using golang and utilizes a PostgreSQL database for storing and indexing data. We operate an instance of the explorer at coda-chain.com. The explorer is licensed under the Apache 2 license.

## Features
### Indexer
* Retrieves latest block data, normalize it and save it to a postgresql db
* Normalizes and saves each transaction included in a block
* Normalizes and saves each snark job included in a block
* Properly handles chain forks & reorganizations
* For each account mutated by a block its latest information (balance, deleagtions) are retreived and saved to the database

### Frontend
1. Summary of most recent blockchain information
    - Latest block number, slot number epoch number
    - Latest canonical snark proof
    - Number of peers, accounts & transactions during the last 24h
    - Number of active validators, total staked currency
    - Number of active snark workers, coda spent on snark work during the last 24h
    - Current inflation, total supply & block reward
2. Historical information & charts
    - Charts on number of blocks produced, peers connected to by node, number of transactions
    - Charts on number of coda staked, number of validators, number of Snark workers, fee breakdown of coda spent on Snark work
    - Number of accounts in the ledger
3. Block viewer
    - Successful blocks
    - Transactions inside each block, type of transactions, coinbase transaction
    - Slot number & epoch number
    - Address of block producer
    - Staged ledger hash
    - Snarked ledger hash
    - Snark jobs and fees included in the block
    - Timestamp
4. Transaction viewer
    - Sender & receiver information
    - Tx hash
    - Fee and amount
    - Nonce & memo
    - Timestamp
5. Account viewer
    - Balance and staked coda information (whether delegated or direct)
    - List of transactions sent or received by account
    - List of any other accounts this account's balance is delegate to
    - List of snark jobs executed by the account, including timestamp, block hash and fees associated for each job
    - List of blocks produced by account, including timestamp and block hash

## Getting started

We currently do not provide any pre-built binaries of the explorer. Docker images are available at https://hub.docker.com/repository/docker/gobitfly/coda-explorer.

- Download the latest version of the coda client and start it with the `-archive` flag set in addition to the currently recommended set of flags
- Wait till the client finishes the initial sync
- Setup a PostgreSQL DB and import the `schema.sql` file from the root of this repository
- Install go version 1.13 or higher
- Clone the repository and run `make all` to build the indexer and front-end binaries
- Start the explorer binary and pass the path to the config file as argument