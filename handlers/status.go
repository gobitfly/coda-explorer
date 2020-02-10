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

var statusTemplate = template.Must(template.New("index").Funcs(templates.GetTemplateFuncs()).ParseFiles("templates/layout.html", "templates/status.html"))

// Status will return the "status" page using a go template
func Status(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	data := &types.PageData{
		Meta: &types.Meta{
			Title:       "coda explorer",
			Description: "",
			Path:        "",
		},
		ShowSyncingMessage: false,
		Active:             "status",
		Version:            version.Version,
	}

	status := &types.DaemonStatus{}
	err := db.DB.Get(status, "SELECT * FROM daemonstatus ORDER BY ts DESC limit 1")
	if err != nil {
		logger.Errorf("error retrieving latest daemon status: %v", err)
		http.Error(w, "Internal server error", 503)
		return
	}
	data.Data = status

	err = statusTemplate.ExecuteTemplate(w, "layout", data)

	if err != nil {
		logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", 503)
		return
	}
}
