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
	"github.com/tankbusta/go-ip2location"
	"time"
)

// PageData is a struct to hold web page data
type PageData struct {
	Active             string
	Meta               *Meta
	ShowSyncingMessage bool
	Data               interface{}
	Version            string
}

// Meta is a struct to hold metadata about the page
type Meta struct {
	Title       string
	Description string
	Path        string
	Tlabel1     string
	Tdata1      string
	Tlabel2     string
	Tdata2      string
}

// IndexPageData is a struct to hold info for the main web page
type IndexPageData struct {
	CurrentEpoch     int      `json:"current_epoch"`
	CurrentSlot      int      `json:"current_slot"`
	CurrentHeight    int      `json:"current_height"`
	ActiveValidators int      `json:"active_validators"`
	ActiveWorkers    int      `json:"active_workers"`
	TotalStaked      int      `json:"total_staked"`
	TotalSupply      int      `json:"total_supply"`
	Peers            int      `json:"peers"`
	Blocks           []*Block `json:"blocks"`
}

// DataTableResponse is a struct to hold data for data table responses
type DataTableResponse struct {
	Draw            int64           `json:"draw"`
	RecordsTotal    int64           `json:"recordsTotal"`
	RecordsFiltered int64           `json:"recordsFiltered"`
	Data            [][]interface{} `json:"data"`
}

// TxPageData is a struct to hold data for transaction page & the transactions table on the account page
type TxPageData struct {
	BlockStateHash string    `db:"blockstatehash"`
	Canonical      bool      `db:"canonical"`
	Index          int       `db:"index"`
	ID             string    `db:"id"`
	Sender         string    `db:"sender"`
	Recipient      string    `db:"recipient"`
	Memo           string    `db:"memo"`
	Fee            int       `db:"fee"`
	Amount         int       `db:"amount"`
	Nonce          int       `db:"nonce"`
	Delegation     bool      `db:"delegation"`
	Ts             time.Time `db:"ts"`
	Slot           int       `db:"slot"`
	Height         int       `db:"height"`
	Epoch          int       `db:"epoch"`
}

// SnarkJobPageData is a struct to hold data for the snarkjob table on the accounts page
type SnarkJobPageData struct {
	BlockStateHash string        `db:"blockstatehash"`
	Canonical      bool          `db:"canonical"`
	Index          int           `db:"index"`
	Jobids         pq.Int64Array `db:"jobids"`
	Prover         string        `db:"prover"`
	Fee            int           `db:"fee"`
	Ts             time.Time     `db:"ts"`
	Slot           int           `db:"slot"`
	Height         int           `db:"height"`
	Epoch          int           `db:"epoch"`
}

type ChartsPageData struct {
	Statistics []*Statistic
	Peers      map[string]*PeerInfoPageData
}

type PeerInfoPageData struct {
	PeerCount int
	Geo       *ip2location.IP2LocationEntry
}
