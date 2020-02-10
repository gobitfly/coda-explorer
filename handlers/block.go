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
	"github.com/gorilla/mux"
	"html/template"
	"net/http"
	"strconv"
)

var blockTemplate = template.Must(template.New("blocks").Funcs(templates.GetTemplateFuncs()).ParseFiles("templates/layout.html", "templates/block.html"))

// Blocks will return information about blocks using a go template
func Block(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "text/html")

	data := &types.PageData{
		Meta: &types.Meta{
			Title:       "coda explorer",
			Description: "",
			Path:        "",
		},
		ShowSyncingMessage: false,
		Active:             "blocks",
		Data:               nil,
		Version:            version.Version,
	}

	vars := mux.Vars(r)
	hash := vars["hash"]

	blockHeight, err := strconv.Atoi(hash)

	var block *types.Block

	if err == nil {
		block, err = db.GetBlockByHeight(blockHeight)
	} else {
		block, err = db.GetBlockByHash(hash)
	}

	if err != nil {
		logger.Errorf("error retrieving block data for block %v: %v", hash, err)
		http.Error(w, "Internal server error", 503)
		return
	}
	data.Data = block

	err = blockTemplate.ExecuteTemplate(w, "layout", data)

	if err != nil {
		logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", 503)
		return
	}
}
