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
	"net/http"
	"strconv"
)

// Search handles search requests
func Search(w http.ResponseWriter, r *http.Request) {

	search := r.FormValue("search")

	_, err := strconv.Atoi(search)

	if err == nil {
		http.Redirect(w, r, "/block/"+search, 301)
		return
	}

	var resCount int
	err = db.DB.Get(&resCount, "SELECT COUNT(*) FROM blocks WHERE statehash = $1", search)
	if resCount > 0 && err == nil {
		http.Redirect(w, r, "/block/"+search, 301)
		return
	}

	err = db.DB.Get(&resCount, "SELECT COUNT(*) FROM userjobs WHERE id = $1", search)
	if resCount > 0 && err == nil {
		http.Redirect(w, r, "/tx/"+search, 301)
		return
	}

	err = db.DB.Get(&resCount, "SELECT COUNT(*) FROM accounts WHERE publickey = $1", search)
	if resCount > 0 && err == nil {
		http.Redirect(w, r, "/account/"+search, 301)
		return
	}

	http.Error(w, "not found", 404)
}
