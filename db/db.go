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

package db

import (
	"coda-explorer/types"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	"time"
)

var logger = logrus.New().WithField("module", "db")

// DB holds the current DB connection
var DB *sqlx.DB

// BlockExists checks if a block is already present in the database
func BlockExists(stateHash string) (bool, error) {
	var stateHashDb string
	err := DB.Get(&stateHashDb, "SELECT statehash FROM blocks WHERE statehash = $1", stateHash)
	return err == nil && stateHashDb == stateHash, err
}

// SaveAccount saves or updates an account in the database
func SaveAccount(account *types.Account) error {
	tx, err := DB.Beginx()

	if err != nil {
		return fmt.Errorf("error starting db tx: %w", err)
	}
	defer tx.Rollback()

	_, err = tx.NamedExec(`INSERT INTO accounts (
								publickey,
								balance,
								nonce,
								receiptchainhash,
								delegate,
								votingfor,
								txsent,
								txreceived,
								blocksproposed,
								snarkjobs,
								firstseen,
								lastseen
							) VALUES (
								:publickey,
								:balance,
								:nonce,
								:receiptchainhash,
								:delegate,
								:votingfor,
								:txsent,
								:txreceived,
								:blocksproposed,
								:snarkjobs,
								:firstseen,
								:lastseen
							) ON CONFLICT (publickey) DO UPDATE SET 
								balance = EXCLUDED.balance, 
								nonce = EXCLUDED.nonce,
								receiptchainhash = EXCLUDED.receiptchainhash,
								delegate = EXCLUDED.delegate,
								votingfor = EXCLUDED.votingfor,
								firstseen = LEAST(EXCLUDED.firstseen, accounts.firstseen),
								lastseen = GREATEST(EXCLUDED.lastseen, accounts.lastseen)
                             `, account)

	if err != nil {
		return fmt.Errorf("error saving account %v db tx: %w", account.PublicKey, err)
	}

	err = tx.Commit()

	return err
}

// SaveBlock saves a new block to the database, checks if the block has already been indexed
func SaveBlock(block *types.Block) error {

	exists, err := BlockExists(block.StateHash)
	if err == nil && exists {
		return fmt.Errorf("error block %v has already been indexed", block.StateHash)
	}

	tx, err := DB.Beginx()

	if err != nil {
		return fmt.Errorf("error starting db tx: %w", err)
	}
	defer tx.Rollback()

	logger.Infof("saving block data")
	_, err = tx.NamedExec(`INSERT INTO blocks (
									statehash,
                    				canonical,
									previousstatehash,
									snarkedledgerhash,
									stagedledgerhash,
									coinbase,
									creator,
									slot,
									height,
									epoch,
									ts,
									totalcurrency,
									usercommandscount,
									snarkjobscount,
									feetransfercount
								) VALUES (
									:statehash,
									:canonical,
									:previousstatehash,
									:snarkedledgerhash,
									:stagedledgerhash,
									:coinbase,
									:creator,
									:slot,
									:height,
									:epoch,
									:ts,
									:totalcurrency,
									:usercommandscount,
									:snarkjobscount,
									:feetransfercount)
								ON CONFLICT DO NOTHING`, block)

	if err != nil {
		return fmt.Errorf("error executing block insert db query: %w", err)
	}

	logger.Infof("saving snark job data")
	for _, sj := range block.SnarkJobs {
		_, err = tx.NamedExec(`INSERT INTO snarkjobs (blockstatehash, index, jobids, prover, fee) VALUES (:blockstatehash, :index, :jobids, :prover, :fee) ON CONFLICT DO NOTHING`, sj)
		if err != nil {
			return fmt.Errorf("error executing snark job insert db query: %w", err)
		}
	}

	logger.Infof("saving fee transfers data")
	for _, ft := range block.FeeTransfers {
		_, err := tx.NamedExec(`INSERT INTO feetransfers (blockstatehash, index, recipient, fee) VALUES (:blockstatehash, :index, :recipient, :fee) ON CONFLICT DO NOTHING `, ft)
		if err != nil {
			return fmt.Errorf("error executing fee transfer insert db query: %w", err)
		}
	}

	logger.Infof("saving user jobs data")
	for _, uj := range block.UserJobs {
		_, err := tx.NamedExec(`INSERT INTO userjobs (blockstatehash, index, id, sender, recipient, memo, fee, amount, nonce, delegation) VALUES (:blockstatehash, :index, :id, :sender, :recipient, :memo, :fee, :amount, :nonce, :delegation) ON CONFLICT DO NOTHING`, uj)
		if err != nil {
			return fmt.Errorf("error executing fee transfer insert db query: %w", err)
		}

	}

	logger.Infof("committing tx")

	err = tx.Commit()
	return err
}

func MarkBlockCanonical(block *types.Block) error {
	tx, err := DB.Beginx()

	if err != nil {
		return fmt.Errorf("error starting db tx: %w", err)
	}
	defer tx.Rollback()

	var canonical bool
	err = tx.Get(&canonical, "SELECT canonical FROM blocks WHERE statehash = $1", block.StateHash)
	if err != nil {
		return fmt.Errorf("error retrieving canonical status from db: %w", err)
	}

	if canonical {
		logger.Infof("block %v at height %v has already been marked as canonical", block.StateHash, block.Height)
		return nil
	}

	_, err = tx.Exec(`UPDATE blocks SET canonical = true WHERE statehash = $1`, block.StateHash)

	if err != nil {
		return fmt.Errorf("error executing block canonical update db query: %w", err)
	}

	logger.Infof("updating snark jobs statistics")
	for _, sj := range block.SnarkJobs {
		_, err := tx.Exec("UPDATE accounts SET snarkjobs = snarkjobs + 1 WHERE publickey = $1", sj.Prover)
		if err != nil {
			return fmt.Errorf("error incrementing snarkjobs column of account table for pk %v: %w", sj.Prover, err)
		}
	}

	logger.Infof("updating user jobs statistics")
	for _, uj := range block.UserJobs {
		_, err = tx.Exec("UPDATE accounts SET txsent = txsent + 1 WHERE publickey = $1", uj.Sender)
		if err != nil {
			return fmt.Errorf("error incrementing txsent column of account table for pk %v: %w", uj.Sender, err)
		}

		_, err = tx.Exec("UPDATE accounts SET txreceived = txreceived + 1 WHERE publickey = $1", uj.Recipient)
		if err != nil {
			return fmt.Errorf("error incrementing txreceived column of account table for pk %v: %w", uj.Recipient, err)
		}
	}

	logger.Infof("updating proposed blocks statistics table")
	_, err = tx.Exec("UPDATE accounts SET blocksproposed = blocksproposed + 1 WHERE publickey = $1", block.Creator)
	if err != nil {
		return fmt.Errorf("error incrementing blocksproposed column of accounts table: %w", err)
	}

	logger.Infof("committing tx")

	err = tx.Commit()
	return err
}

func MarkBlockOrphaned(block *types.Block) error {
	tx, err := DB.Beginx()

	if err != nil {
		return fmt.Errorf("error starting db tx: %w", err)
	}
	defer tx.Rollback()

	var canonical bool
	err = tx.Get(&canonical, "SELECT canonical FROM blocks WHERE statehash = $1", block.StateHash)
	if err != nil {
		return fmt.Errorf("error retrieving canonical status from db: %w", err)
	}

	if !canonical {
		logger.Infof("block %v at height %v has already been marked as orphaned", block.StateHash, block.Height)
		return nil
	}

	_, err = tx.Exec(`UPDATE blocks SET canonical = false WHERE statehash = $1`, block.StateHash)

	if err != nil {
		return fmt.Errorf("error executing block canonical update db query: %w", err)
	}

	logger.Infof("updating snark jobs statistics")
	for _, sj := range block.SnarkJobs {
		_, err := tx.Exec("UPDATE accounts SET snarkjobs = snarkjobs - 1 WHERE publickey = $1", sj.Prover)
		if err != nil {
			return fmt.Errorf("error incrementing snarkjobs column of account table for pk %v: %w", sj.Prover, err)
		}
	}

	logger.Infof("updating user jobs statistics")
	for _, uj := range block.UserJobs {
		_, err = tx.Exec("UPDATE accounts SET txsent = txsent - 1 WHERE publickey = $1", uj.Sender)
		if err != nil {
			return fmt.Errorf("error incrementing txsent column of account table for pk %v: %w", uj.Sender, err)
		}

		_, err = tx.Exec("UPDATE accounts SET txreceived = txreceived - 1 WHERE publickey = $1", uj.Recipient)
		if err != nil {
			return fmt.Errorf("error incrementing txreceived column of account table for pk %v: %w", uj.Recipient, err)
		}
	}

	logger.Infof("committing tx")

	err = tx.Commit()
	return err
}

// RollbackBlock removes a block from the database, rolling back all mutations to the account counters
func RollbackBlock(block *types.Block) error {
	tx, err := DB.Begin()

	if err != nil {
		return fmt.Errorf("error starting db tx: %w", err)
	}
	defer tx.Rollback()

	_, err = tx.Exec("UPDATE accounts SET blocksproposed = blocksproposed - 1 WHERE publickey = $1", block.Creator)
	if err != nil {
		return fmt.Errorf("error decrementing blocksporposed column of account table for pk %v: %w", block.Creator, err)
	}

	for _, snarkJob := range block.SnarkJobs {
		_, err := tx.Exec("UPDATE accounts SET snarkjobs = snarkjobs - 1 WHERE publickey = $1", snarkJob.Prover)
		if err != nil {
			return fmt.Errorf("error decrementing snarkjobs column of account table for pk %v: %w", snarkJob.Prover, err)
		}
	}
	for _, userJob := range block.UserJobs {
		_, err := tx.Exec("UPDATE accounts SET txsent = txsent - 1 WHERE publickey = $1", userJob.Sender)
		if err != nil {
			return fmt.Errorf("error decrementing txsent column of account table for pk %v: %w", userJob.Sender, err)
		}

		_, err = tx.Exec("UPDATE accounts SET txreceived = txreceived - 1 WHERE publickey = $1", userJob.Recipient)
		if err != nil {
			return fmt.Errorf("error decrementing txreceived column of account table for pk %v: %w", userJob.Recipient, err)
		}

		_, err = tx.Exec("DELETE FROM accounttransactions WHERE id = $1", userJob.ID)
		if err != nil {
			return fmt.Errorf("error deleting job %v from accounttransactions table: %w", userJob.ID, err)
		}
	}

	_, err = tx.Exec("DELETE FROM snarkjobs WHERE blockstatehash = $1", block.StateHash)
	if err != nil {
		return fmt.Errorf("error block %v from snarkjobs table: %w", block.StateHash, err)
	}

	_, err = tx.Exec("DELETE FROM feetransfers WHERE blockstatehash = $1", block.StateHash)
	if err != nil {
		return fmt.Errorf("error block %v from feetransfers table: %w", block.StateHash, err)
	}

	_, err = tx.Exec("DELETE FROM userjobs WHERE blockstatehash = $1", block.StateHash)
	if err != nil {
		return fmt.Errorf("error block %v from userjobs table: %w", block.StateHash, err)
	}

	_, err = tx.Exec("DELETE FROM blocks WHERE statehash = $1", block.StateHash)
	if err != nil {
		return fmt.Errorf("error deleting block %v from blocks table: %w", block.StateHash, err)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("error committing block %v rollback transaction: %w", block.StateHash, err)
	}

	return nil
}

// GetBlockByHeight retrieves a block from the database by its canonical height
func GetBlockByHeight(height int) (*types.Block, error) {
	var stateHash string
	err := DB.Get(&stateHash, "SELECT statehash FROM blocks WHERE height = $1", height)

	if err != nil {
		return nil, fmt.Errorf("error block at height %v not found: %w", height, err)
	}

	return GetBlockByHash(stateHash)
}

// GetLastBlockHashes retrieves a set of blocks from the database by their canonical height
func GetLastBlockHashes(lookback int) ([]*types.BlockHashNumber, error) {
	var hashes []*types.BlockHashNumber
	err := DB.Select(&hashes, "SELECT statehash, canonical, previousstatehash, height FROM blocks ORDER BY height DESC limit $1", lookback)

	if err != nil {
		return nil, fmt.Errorf("error retrieving last block hashes: %w", err)
	}

	return hashes, nil
}

// GetBlockByHash retrieves a block from the database by its canonical state hash
func GetBlockByHash(hash string) (*types.Block, error) {
	block := &types.Block{
		SnarkJobs:    []*types.SnarkJob{},
		FeeTransfers: []*types.FeeTransfer{},
		UserJobs:     []*types.UserJob{},
	}

	err := DB.Get(block, "SELECT * FROM blocks WHERE statehash = $1", hash)

	if err != nil {
		return nil, fmt.Errorf("error retrieving data for block %v from the database: %w", hash, err)
	}

	if block.SnarkJobsCount > 0 {
		err = DB.Select(&block.SnarkJobs, "SELECT * FROM snarkjobs WHERE blockstatehash = $1 ORDER BY index", hash)
		if err != nil {
			return nil, fmt.Errorf("error retrieving snark job data for block %v from the database: %w", hash, err)
		}
	}

	if block.FeeTransferCount > 0 {
		err = DB.Select(&block.FeeTransfers, "SELECT * FROM feetransfers WHERE blockstatehash = $1 ORDER BY index", hash)
		if err != nil {
			return nil, fmt.Errorf("error retrieving fee transfer data for block %v from the database: %w", hash, err)
		}
	}

	if block.UserCommandsCount > 0 {
		err = DB.Select(&block.UserJobs, "SELECT * FROM userjobs WHERE blockstatehash = $1 ORDER BY index", hash)
		if err != nil {
			return nil, fmt.Errorf("error retrieving user jobs data for block %v from the database: %w", hash, err)
		}
	}

	return block, nil
}

// SaveDaemonStatus saves the daemon status the the database
func SaveDaemonStatus(daemonStatus *types.DaemonStatus) error {
	_, err := DB.NamedExec(`INSERT INTO daemonstatus (
						  ts,
                          blockchainlength,
                          commitid,
                          epochduration,
                          slotduration,
                          slotsperepoch,
                          consensusmechanism,
                          highestblocklengthreceived,
                          ledgermerkleroot,
                          numaccounts,
                          peers,
                          peerscount,
                          statehash,
                          syncstatus,
                          uptime
						) VALUES (
						  :ts,
                          :blockchainlength,
                          :commitid,
                          :epochduration,
                          :slotduration,
                          :slotsperepoch,
                          :consensusmechanism,
                          :highestblocklengthreceived,
                          :ledgermerkleroot,
                          :numaccounts,
                          :peers,
                          :peerscount,
                          :statehash,
                          :syncstatus,
                          :uptime
						) ON CONFLICT DO NOTHING`, daemonStatus)
	if err != nil {
		return fmt.Errorf("error saving daemon status: %w", err)
	}

	return nil
}

// GenerateAndSaveStatistics generates the statistics for a given day and saves them to the database
func GenerateAndSaveStatistics(date time.Time) error {
	startDate := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
	endDate := time.Date(date.Year(), date.Month(), date.Day(), 23, 59, 59, 0, time.UTC)

	logger.Infof("processing statistics for day %v", startDate)
	tx, err := DB.Begin()

	if err != nil {
		return fmt.Errorf("error starting db tx: %w", err)
	}
	defer tx.Rollback()

	// Number of daily blocks produced
	indicator := "BLOCK_COUNT"
	_, err = tx.Exec(`INSERT INTO statistics (indicator, ts, value) SELECT $1, $2, COUNT(*) FROM blocks WHERE ts >= $2 AND ts <= $3  ON CONFLICT (indicator, ts) DO UPDATE SET value = EXCLUDED.value;`, indicator, startDate, endDate)
	if err != nil {
		return fmt.Errorf("error executing %s statistics query for day %v: %w", indicator, startDate, err)
	}

	// Number of daily tx
	indicator = "TX_COUNT"
	_, err = tx.Exec(`INSERT INTO statistics (indicator, ts, value) SELECT $1, $2, COALESCE(SUM(usercommandscount), 0) FROM blocks WHERE ts >= $2 AND ts <= $3  ON CONFLICT (indicator, ts) DO UPDATE SET value = EXCLUDED.value;`, indicator, startDate, endDate)
	if err != nil {
		return fmt.Errorf("error executing %s statistics query for day %v: %w", indicator, startDate, err)
	}

	// Total supply
	indicator = "TOTAL_SUPPLY"
	_, err = tx.Exec(`INSERT INTO statistics (indicator, ts, value) SELECT $1, $2, COALESCE(MAX(totalcurrency), 0) FROM blocks WHERE ts >= $2 AND ts <= $3  ON CONFLICT (indicator, ts) DO UPDATE SET value = EXCLUDED.value;`, indicator, startDate, endDate)
	if err != nil {
		return fmt.Errorf("error executing %s statistics query for day %v: %w", indicator, startDate, err)
	}

	// Number of daily active block producers
	indicator = "BLOCK_PRODUCERS"
	_, err = tx.Exec(`INSERT INTO statistics (indicator, ts, value) SELECT $1, $2, COUNT(DISTINCT creator) FROM blocks WHERE ts >= $2 AND ts <= $3  ON CONFLICT (indicator, ts) DO UPDATE SET value = EXCLUDED.value;`, indicator, startDate, endDate)
	if err != nil {
		return fmt.Errorf("error executing %s statistics query for day %v: %w", indicator, startDate, err)
	}

	// Number of daily new accounts
	indicator = "NEW_ACCOUNTS"
	_, err = tx.Exec(`INSERT INTO statistics (indicator, ts, value) SELECT $1, $2, COUNT(*) FROM accounts WHERE firstseen >= $2 AND firstseen <= $3  ON CONFLICT (indicator, ts) DO UPDATE SET value = EXCLUDED.value;`, indicator, startDate, endDate)
	if err != nil {
		return fmt.Errorf("error executing %s statistics query for day %v: %w", indicator, startDate, err)
	}

	// Number of daily active snark workers
	indicator = "SNARK_WORKERS"
	_, err = tx.Exec(`INSERT INTO statistics (indicator, ts, value) SELECT $1, $2, COUNT(DISTINCT prover) FROM snarkjobs LEFT JOIN blocks ON blocks.statehash = snarkjobs.blockstatehash WHERE blocks.ts >= $2 AND blocks.ts <= $3  ON CONFLICT (indicator, ts) DO UPDATE SET value = EXCLUDED.value;`, indicator, startDate, endDate)
	if err != nil {
		return fmt.Errorf("error executing %s statistics query for day %v: %w", indicator, startDate, err)
	}

	// Number of daily coins spent on snarks
	indicator = "SNARK_FEES"
	_, err = tx.Exec(`INSERT INTO statistics (indicator, ts, value) SELECT $1, $2, COALESCE(SUM(fee), 0) FROM snarkjobs LEFT JOIN blocks ON blocks.statehash = snarkjobs.blockstatehash WHERE blocks.ts >= $2 AND blocks.ts <= $3  ON CONFLICT (indicator, ts) DO UPDATE SET value = EXCLUDED.value;`, indicator, startDate, endDate)
	if err != nil {
		return fmt.Errorf("error executing %s statistics query for day %v: %w", indicator, startDate, err)
	}

	// Number of daily seen unique peers
	indicator = "PEERS"
	_, err = tx.Exec(`INSERT INTO statistics (indicator, ts, value) (SELECT $1, $2, COUNT(DISTINCT peer) FROM (SELECT UNNEST(peers) AS peer FROM daemonstatus WHERE ts >= $2 AND ts <= $3) AS a) ON CONFLICT (indicator, ts) DO UPDATE SET value = EXCLUDED.value;`, indicator, startDate, endDate)
	if err != nil {
		return fmt.Errorf("error executing %s statistics query for day %v: %w", indicator, startDate, err)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("error committing statistics transaction for day %v: %w", startDate, err)
	}
	logger.Infof("statistics for day %v generated & saved", startDate)
	return nil
}
