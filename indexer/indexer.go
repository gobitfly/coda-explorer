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
	"coda-explorer/types"
	"fmt"
	"github.com/sirupsen/logrus"
	"sync"
	"time"
)

var logger = logrus.New().WithField("module", "indexer")

// Start starts the indexing process
func Start(rpcEndpoint string) {
	client := rpc.NewCodaClient(rpcEndpoint)

	newBlockChan := make(chan string)

	go client.WatchNewBlocks(newBlockChan)

	go exportDaemonStatus(client, time.Minute*10)

	go checkNewBlocks(newBlockChan, client, time.Minute)

	go updateStatistics(time.Hour)

	checkBlocks(client, 1000)
}

func updateStatistics(intv time.Duration) {
	ticker := time.NewTicker(intv)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			err := db.GenerateAndSaveStatistics(time.Now().Add(time.Hour * 24 * -1))
			if err != nil {
				logger.Errorf("error generating statistics: %w", err)
			}
		}
	}
}

// Periodically checks for forked or missing blocks
func checkNewBlocks(newBlockChan chan string, client *rpc.CodaClient, intv time.Duration) {
	ticker := time.NewTicker(intv)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			checkBlocks(client, 10)
		case <-newBlockChan:
			checkBlocks(client, 10)
		}
	}
}

var checkBlockMux = &sync.Mutex{}

func checkBlocks(client *rpc.CodaClient, lookback int) {
	checkBlockMux.Lock()
	defer checkBlockMux.Unlock()

	dbBlocks, err := db.GetLastBlockHashes(lookback)
	if err != nil {
		logger.Errorf("error retrieving last %v blocks from the databases: %w", lookback, err)
		return
	}
	dbBlocksMap := make(map[string]bool)
	for _, b := range dbBlocks {
		dbBlocksMap[b.StateHash] = true
	}

	nodeBlocks, err := client.GetLastBlocks(lookback)
	if err != nil {
		logger.Errorf("error retrieving last %v blocks from the rpc node: %w", lookback, err)
		return
	}
	for _, b := range nodeBlocks {
		_, present := dbBlocksMap[b.StateHash]
		if present {
			// Block has already been properly indexed
			continue
		} else {
			err := exportBlock(b, client)
			if err != nil {
				logger.Errorf("error exporting block %v at height %v: %w", b.StateHash, b.Height, err)
			}
		}
	}

	dbBlocks, err = db.GetLastBlockHashes(lookback)
	if err != nil {
		logger.Errorf("error retrieving last %v blocks from the databases: %w", lookback, err)
		return
	}

	currentHash := ""

	for i, block := range dbBlocks {
		if i == 0 {
			currentHash = block.PreviousStateHash

			if !block.Canonical {
				blockData, err := db.GetBlockByHash(block.StateHash)
				if err != nil {
					logger.Errorf("error retrieving data for block %v at height %v: %w", block.StateHash, block.Height, err)
					return
				}

				err = db.MarkBlockCanonical(blockData)
				if err != nil {
					logger.Errorf("error marking block %v at height %v as canonical: %w", block.StateHash, block.Height, err)
					return
				}
			}
		} else {
			if block.StateHash == currentHash && !block.Canonical { // block is part of the canonical chain but currently not marked as canonical
				logger.Infof("marking block %v at height %v as canonical", block.StateHash, block.Height)
				blockData, err := db.GetBlockByHash(block.StateHash)
				if err != nil {
					logger.Errorf("error retrieving data for block %v at height %v: %w", block.StateHash, block.Height, err)
					return
				}

				err = db.MarkBlockCanonical(blockData)
				if err != nil {
					logger.Errorf("error marking block %v at height %v as canonical: %w", block.StateHash, block.Height, err)
					return
				}
				currentHash = block.PreviousStateHash
			} else if block.StateHash != currentHash && block.Canonical { // block is not part of the canonical chain but currently marked as canonical
				logger.Infof("marking block %v at height %v as orphaned", block.StateHash, block.Height)
				blockData, err := db.GetBlockByHash(block.StateHash)
				if err != nil {
					logger.Errorf("error retrieving data for block %v at height %v: %w", block.StateHash, block.Height, err)
					return
				}

				err = db.MarkBlockOrphaned(blockData)
				if err != nil {
					logger.Errorf("error marking block %v at height %v as canonical: %w", block.StateHash, block.Height, err)
					return
				}
			} else if block.Canonical {
				currentHash = block.PreviousStateHash
			}
		}
	}

}

// Exports a block to the database, does nothing if the block has already previously been exported
func exportBlock(block *types.Block, client *rpc.CodaClient) error {
	logger.Infof("exporting block %v at height %v", block.StateHash, block.Height)
	exists, err := db.BlockExists(block.StateHash)

	if err == nil && exists {
		logger.Infof("block %v already exported", block.StateHash)
		return nil
	}

	start := time.Now()

	accountsInBlock := make(map[string]bool)
	accountsInBlock[block.Creator] = true

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
		//logger.Infof("exporting account %v", pubKey)
		account, err := client.GetAccount(pubKey)
		if err != nil {
			return fmt.Errorf("error retrieving account data for account %v via rpc: %w", pubKey, err)
		}
		//logger.Infof("account data retrieved")

		account.FirstSeen = block.Ts
		account.LastSeen = block.Ts

		err = db.SaveAccount(account)
		if err != nil {
			return fmt.Errorf("error saving account data for account %v: %w", pubKey, err)
		}
		//logger.Infof("account data exported to db")
	}
	logger.Infof("accounts updated, saving block to db")

	err = db.SaveBlock(block)
	if err != nil {
		return fmt.Errorf("error saving block data for block %v: %w", block.StateHash, err)
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
