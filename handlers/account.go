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

package handlers

import (
	"coda-explorer/db"
	"coda-explorer/templates"
	"coda-explorer/types"
	"coda-explorer/version"
	"encoding/json"
	"github.com/gorilla/mux"
	"html/template"
	"net/http"
	"strconv"
)

var accountTemplate = template.Must(template.New("blocks").Funcs(templates.GetTemplateFuncs()).ParseFiles("templates/layout.html", "templates/account.html"))

// Blocks will return information about blocks using a go template
func Account(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "text/html")

	data := &types.PageData{
		Meta: &types.Meta{
			Title:       "coda explorer",
			Description: "",
			Path:        "",
		},
		ShowSyncingMessage: false,
		Active:             "accounts",
		Data:               nil,
		Version:            version.Version,
	}

	vars := mux.Vars(r)
	pk := vars["pk"]
	account := &types.Account{}

	err := db.DB.Get(account, "SELECT * FROM accounts WHERE publickey = $1", pk)
	if err != nil {
		logger.Errorf("error retrieving account data for account %v: %v", pk, err)
		http.Error(w, "Internal server error", 503)
		return
	}
	data.Data = account

	err = accountTemplate.ExecuteTemplate(w, "layout", data)

	if err != nil {
		logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", 503)
		return
	}
}

// BlocksData will return information about blocks
func AccountBlocksData(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	q := r.URL.Query()

	draw, err := strconv.ParseInt(q.Get("draw"), 10, 64)
	if err != nil {
		logger.Errorf("error converting datatables data parameter from string to int: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}
	start, err := strconv.ParseInt(q.Get("start"), 10, 64)
	if err != nil {
		logger.Errorf("error converting datatables start parameter from string to int: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}
	length, err := strconv.ParseInt(q.Get("length"), 10, 64)
	if err != nil {
		logger.Errorf("error converting datatables length parameter from string to int: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}
	if length > 100 {
		length = 100
	}

	vars := mux.Vars(r)
	pk := vars["pk"]

	var blocksCount int64

	err = db.DB.Get(&blocksCount, "SELECT blocksproposed FROM accounts WHERE publickey = $1", pk)
	if err != nil {
		logger.Errorf("error retrieving blockproposed for account %v: %v", pk, err)
		http.Error(w, "Internal server error", 503)
		return
	}
	if blocksCount > 10000 {
		blocksCount = 10000
	}

	var blocks []*types.Block

	err = db.DB.Select(&blocks, `SELECT *
										FROM blocks 
										WHERE creator = $1
										ORDER BY blocks.height DESC, canonical DESC LIMIT $2 OFFSET $3`, pk, length, start)

	if err != nil {
		logger.Errorf("error retrieving block data for account %v: %v", pk, err)
		http.Error(w, "Internal server error", 503)
		return
	}

	tableData := make([][]interface{}, len(blocks))
	for i, b := range blocks {
		tableData[i] = []interface{}{
			b.Canonical,
			b.Height,
			b.Epoch,
			b.Slot,
			b.Ts.Unix(),
			b.Creator,
			b.StateHash,
			b.UserCommandsCount,
			b.SnarkJobsCount,
			b.Coinbase,
		}
	}

	data := &types.DataTableResponse{
		Draw:            draw,
		RecordsTotal:    blocksCount,
		RecordsFiltered: blocksCount,
		Data:            tableData,
	}

	err = json.NewEncoder(w).Encode(data)
	if err != nil {
		logger.Errorf("error enconding json response for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", 503)
		return
	}
}
