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

package types

import (
	"github.com/lib/pq"
	"time"
)

// Block represents a row of the blocks db table
type Block struct {
	StateHash         string    `db:"statehash"`
	PreviousStateHash string    `db:"previousstatehash"`
	SnarkedLedgerHash string    `db:"snarkedledgerhash"`
	StagedLedgerHash  string    `db:"stagedledgerhash"`
	Coinbase          int       `db:"coinbase"`
	Creator           string    `db:"creator"`
	Slot              int       `db:"slot"`
	Height            int       `db:"height"`
	Epoch             int       `db:"epoch"`
	Ts                time.Time `db:"ts"`
	TotalCurrency     int       `db:"totalcurrency"`
	UserCommandsCount int       `db:"usercommandscount"`
	SnarkJobsCount    int       `db:"snarkjobscount"`
	FeeTransferCount  int       `db:"feetransfercount"`

	SnarkJobs    []*SnarkJob
	FeeTransfers []*FeeTransfer
	UserJobs     []*UserJob
}

// BlockHashNumber is a helper type that contains only the hash and height of a block
type BlockHashNumber struct {
	StateHash string `db:"statehash"`
	Height    int    `db:"height"`
}

// SnarkJob represents a row of the snarkjobs db table
type SnarkJob struct {
	BlockStateHash string        `db:"blockstatehash"`
	Index          int           `db:"index"`
	Jobids         pq.Int64Array `db:"jobids"`
	Prover         string        `db:"prover"`
	Fee            int           `db:"fee"`
}

// FeeTransfer represents a row of the feetransfers db table
type FeeTransfer struct {
	BlockStateHash string `db:"blockstatehash"`
	Index          int    `db:"index"`
	Recipient      string `db:"recipient"`
	Fee            int    `db:"fee"`
}

// UserJob represents a row of the userjobs db table
type UserJob struct {
	BlockStateHash string `db:"blockstatehash"`
	Index          int    `db:"index"`
	ID             string `db:"id"`
	Sender         string `db:"sender"`
	Recipient      string `db:"recipient"`
	Memo           string `db:"memo"`
	Fee            int    `db:"fee"`
	Amount         int    `db:"amount"`
	Nonce          int    `db:"nonce"`
	Delegation     bool   `db:"delegation"`
}

// Account represents a row of the accounts db table
type Account struct {
	PublicKey        string    `db:"publickey"`
	Balance          int       `db:"balance"`
	Nonce            int       `db:"nonce"`
	ReceiptChainHash string    `db:"receiptchainhash"`
	Delegate         string    `db:"delegate"`
	VotingFor        string    `db:"votingfor"`
	TxSent           int       `db:"txsent"`
	TxReceived       int       `db:"txreceived"`
	BlocksProposed   int       `db:"blocksproposed"`
	SnarkJobs        int       `db:"snarkjobs"`
	FirstSeen        time.Time `db:"firstseen"`
	LastSeen         time.Time `db:"lastseen"`
}

// AccountTransaction represents a row of the accounttransactions db table
type AccountTransaction struct {
	PublicKey string    `db:"publickey"`
	ID        string    `db:"id"`
	Ts        time.Time `db:"ts"`
}

// DaemonStatus represents a row of the daemonstatus db table
type DaemonStatus struct {
	Ts                         time.Time      `db:"ts"`
	BlockchainLength           int            `db:"blockchainlength"`
	CommitID                   string         `db:"commitid"`
	EpochDuration              int            `db:"epochduration"`
	SlotDuration               int            `db:"slotduration"`
	SlotsPerEpoch              int            `db:"slotsperepoch"`
	ConsensusMechanism         string         `db:"consensusmechanism"`
	HighestBlockLengthReceived int            `db:"highestblocklengthreceived"`
	LedgerMerkleRoot           string         `db:"ledgermerkleroot"`
	NumAccounts                int            `db:"numaccounts"`
	Peers                      pq.StringArray `db:"peers"`
	PeersCount                 int            `db:"peerscount"`
	StateHash                  string         `db:"statehash"`
	SyncStatus                 string         `db:"syncstatus"`
	Uptime                     int            `db:"uptime"`
}

// Statistic represents a row of the statistics db table
type Statistic struct {
	Indicator string    `db:"indicator"`
	Ts        time.Time `db:"ts"`
	Value     float64   `db:"value"`
}