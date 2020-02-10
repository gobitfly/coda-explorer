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

package services

import (
	"coda-explorer/db"
	"coda-explorer/types"
	"fmt"
	"github.com/sirupsen/logrus"
	"sync"
	"sync/atomic"
	"time"
)

var latestHeight uint64
var indexPageData atomic.Value
var ready = sync.WaitGroup{}

var logger = logrus.New().WithField("module", "services")

// Init will initialize the services
func Init() {
	ready.Add(2)
	go heightUpdater()
	go indexPageDataUpdater()
	ready.Wait()
}

func heightUpdater() {
	firstRun := true

	for true {
		var height uint64
		err := db.DB.Get(&height, "SELECT COALESCE(MAX(height), 0) FROM blocks")

		if err != nil {
			logger.Printf("error retrieving latest height from the database: %v", err)
		} else {
			atomic.StoreUint64(&latestHeight, height)
			if firstRun {
				ready.Done()
				firstRun = false
			}
		}
		time.Sleep(time.Second)
	}
}

func indexPageDataUpdater() {
	firstRun := true

	for true {
		data, err := getIndexPageData()
		if err != nil {
			logger.Errorf("error retrieving index page data: %v", err)
			time.Sleep(time.Second * 10)
			continue
		}
		indexPageData.Store(data)
		if firstRun {
			ready.Done()
			firstRun = false
		}
		time.Sleep(time.Second * 10)
	}
}

func getIndexPageData() (*types.IndexPageData, error) {
	data := &types.IndexPageData{}

	var blocks []*types.Block

	err := db.DB.Select(&blocks, `SELECT *
										FROM blocks 
										ORDER BY blocks.height DESC, canonical DESC LIMIT 20`)

	if err != nil {
		return nil, fmt.Errorf("error retrieving index block data: %w", err)
	}
	data.Blocks = blocks

	if len(blocks) > 0 {
		data.CurrentSlot = blocks[0].Slot
		data.CurrentEpoch = blocks[0].Epoch
		data.CurrentHeight = blocks[0].Height
		data.TotalSupply = blocks[0].TotalCurrency
	}

	err = db.DB.Get(&data.ActiveWorkers, "SELECT COUNT(DISTINCT prover) FROM snarkjobs LEFT JOIN blocks ON blocks.statehash = snarkjobs.blockstatehash WHERE blocks.ts > now() - interval '1 day'")
	if err != nil {
		return nil, fmt.Errorf("error retrieving active workers data: %w", err)
	}

	err = db.DB.Get(&data.ActiveValidators, "select count (distinct creator) from blocks where ts > now() - interval '1 day';")
	if err != nil {
		return nil, fmt.Errorf("error retrieving active validators data: %w", err)
	}

	err = db.DB.Get(&data.TotalStaked, "select COALESCE(sum(balance), 0) from accounts where publickey in (select distinct creator from blocks where ts > now() - interval '1 day');")
	if err != nil {
		return nil, fmt.Errorf("error retrieving total staked data: %w", err)
	}

	return data, nil
}

// LatestIndexPageData returns the latest index page data
func LatestIndexPageData() *types.IndexPageData {
	return indexPageData.Load().(*types.IndexPageData)
}
