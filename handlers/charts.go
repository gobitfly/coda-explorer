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
	"html/template"
	"net/http"
)

// ChartBlocks will return information about the daily produced blocks using a go template
var chartsTemplate = template.Must(template.New("blocks").Funcs(templates.GetTemplateFuncs()).ParseFiles("templates/layout.html", "templates/charts.html"))

func Charts(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "text/html")

	var stats []*types.Statistic
	err := db.DB.Select(&stats, "SELECT * FROM statistics WHERE value > 0 ORDER BY ts, indicator")
	if err != nil {
		logger.Errorf("error retrieving statistcs data for route %v: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", 503)
		return
	}

	data := &types.PageData{
		Meta: &types.Meta{
			Title:       "coda explorer",
			Description: "",
			Path:        "",
		},
		ShowSyncingMessage: false,
		Active:             "charts",
		Data:               stats,
		Version:            version.Version,
	}

	err = chartsTemplate.ExecuteTemplate(w, "layout", data)

	if err != nil {
		logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", 503)
		return
	}
}