/*
 *    Copyright 2020 bitfly gmbh
 *
 *    Licensed under the Apache License, Version 2.0 (the "License");
 *    you may not use this file except in compliance with the License.
 *    You may obtain a copy of the License at
 *
 *        http://www.apache.org/licenses/LICENSE-2.0
 *
 *    Unless required by applicable law or agreed to in writing, software
 *    distributed under the License is distributed on an "AS IS" BASIS,
 *    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *    See the License for the specific language governing permissions and
 *    limitations under the License.
 */

package indexer

import (
	"coda-explorer/db"
	"coda-explorer/rpc"
	"fmt"
	"github.com/sirupsen/logrus"
	"time"
)

var logger = logrus.New().WithField("module", "indexer")

// Start starts the indexing process
func Start(rpcEndpoint string) {
	client := rpc.NewCodaClient(rpcEndpoint)

	newBlockChan := make(chan string)

	go client.WatchNewBlocks(newBlockChan)
	go handleNewBlocks(newBlockChan, client)

	go exportDaemonStatus(client, time.Minute*10)

	go forkChecker(client, time.Minute)
}

// Periodically checks for forked or missing blocks
func forkChecker(client *rpc.CodaClient, intv time.Duration) {
	ticker := time.NewTicker(intv)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			dbBlocks, err := db.GetLastBlockHashes(10)
			if err != nil {
				logger.Errorf("error retrieving last 10 blocks from the databases: %w", err)
				continue
			}
			dbBlocksMap := make(map[int]string)
			for _, b := range dbBlocks {
				dbBlocksMap[b.Height] = b.StateHash
			}

			nodeBlocks, err := client.GetLastBlocks(10)
			if err != nil {
				logger.Errorf("error retrieving last 10 blocks from the rpc node: %w", err)
				continue
			}
			for _, b := range nodeBlocks {
				hash, present := dbBlocksMap[b.Height]
				if present && hash == b.StateHash {
					// Block has already been properly indexed
					continue
				} else if present && hash != b.StateHash {
					// Node block at given height is different to the block stored in the db for the same height
					// Roll back the block present in the db and save the new block
					logger.Infof("block at height %v stored in db is different to block at height %v in the node: %v != %v", b.Height, b.Height, hash, b.StateHash)
					logger.Infof("rolling back block %v at height %v", hash, b.Height)
					orphanedBlock, err := db.GetBlockByHash(hash)
					if err != nil {
						logger.Errorf("error retrieving orphaned block %v at height %v from the db: %w", hash, b.Height, err)
						continue
					}
					err = db.RollbackBlock(orphanedBlock)
					if err != nil {
						logger.Errorf("error rolling back orphaned block %v at height %v: %w", hash, b.Height, err)
						continue
					}
					err = exportBlock(b.StateHash, client)
					if err != nil {
						logger.Errorf("error exporting block %v at height %v: %w", b.StateHash, b.Height, err)
					}
				} else if !present {
					err := exportBlock(b.StateHash, client)
					if err != nil {
						logger.Errorf("error exporting block %v at height %v: %w", b.StateHash, b.Height, err)
					}
				}
			}
		}
	}
}

// Listens for new block events and exports any received new block
func handleNewBlocks(newBlockChan chan string, client *rpc.CodaClient) {
	for {
		select {
		case newBlock := <-newBlockChan:
			err := exportBlock(newBlock, client)
			if err != nil {
				logger.Errorf("error exporting new block %v: %w", newBlock, err)
			}
		}
	}
}

// Exports a block to the database, does nothing if the block has already previously been exported
func exportBlock(stateHash string, client *rpc.CodaClient) error {
	logger.Infof("exporting block %v", stateHash)
	exists, err := db.BlockExists(stateHash)

	if err == nil && exists {
		logger.Infof("block %v already exported", stateHash)
		return nil
	}

	start := time.Now()

	block, err := client.GetBlock(stateHash)
	if err != nil {
		return fmt.Errorf("error retrieving block data for block %v via rpc: %w", stateHash, err)
	}
	logger.Infof("block data received")

	accountsInBlock := make(map[string]bool)
	accountsInBlock[block.Creator] = true
	logger.Println(block.Creator)

	for _, uj := range block.UserJobs {
		accountsInBlock[uj.Sender] = true
		accountsInBlock[uj.Recipient] = true
	}

	for _, ft := range block.FeeTransfers {
		accountsInBlock[ft.Recipient] = true
	}

	for _, sj := range block.SnarkJobs {
		accountsInBlock[sj.Prover] = true
	}
	logger.Infof("block mutated %v accounts", len(accountsInBlock))

	for pubKey := range accountsInBlock {
		logger.Infof("exporting account %v", pubKey)
		account, err := client.GetAccount(pubKey)
		if err != nil {
			return fmt.Errorf("error retrieving account data for account %v via rpc: %w", pubKey, err)
		}
		logger.Infof("account data retrieved")

		account.FirstSeen = block.Ts
		account.LastSeen = block.Ts

		err = db.SaveAccount(account)
		if err != nil {
			return fmt.Errorf("error saving account data for account %v: %w", pubKey, err)
		}
		logger.Infof("account data exported to db")
	}

	err = db.SaveBlock(block)
	if err != nil {
		return fmt.Errorf("error saving block data for block %v: %w", stateHash, err)
	}
	logger.WithField("txs", block.UserCommandsCount).WithField("snarks", block.SnarkJobsCount).WithField("feeTransfers", block.FeeTransferCount).Infof("block data exported to db, took %v", time.Since(start))

	return nil
}

// Exports the current daemon status in a specified interval to the database
func exportDaemonStatus(client *rpc.CodaClient, intv time.Duration) {
	ticker := time.NewTicker(intv)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			status, err := client.GetDaemonStatus()
			if err != nil {
				logger.Errorf("error retrieving daemon status: %w", err)
				continue
			}

			err = db.SaveDaemonStatus(status)
			if err != nil {
				logger.Errorf("error saving daemon status: %w", err)
				continue
			}
			logger.Infof("daemon status updated")
		}
	}
}
