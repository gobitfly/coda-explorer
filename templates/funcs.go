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

package templates

import (
	"fmt"
	"github.com/lib/pq"
	"html/template"
	"strings"
	"time"

	"github.com/akamensky/base58"
	"github.com/leekchan/gtf"
)

// GetTemplateFuncs returns helper functions used in the html templates
func GetTemplateFuncs() template.FuncMap {
	fm := template.FuncMap{
		"formatSeconds":      formatSeconds,
		"formatMilliSeconds": formatMilliSeconds,
		"formatPGIntArray":   formatPGIntArray,
		"decodeBase58":       decodeBase58,
	}

	gtf.ForceInject(fm)
	return fm
}

// Formats a duration in seconds
func formatSeconds(seconds int) string {
	return fmt.Sprintf("%v", time.Second*time.Duration(seconds))
}

// Formats a duration in milliseconds
func formatMilliSeconds(ms int) string {
	return fmt.Sprintf("%v", time.Millisecond*time.Duration(ms))
}

// Formats a array of int64 values (postgresql driver format)
func formatPGIntArray(arr pq.Int64Array) string {
	return strings.Trim(strings.Replace(fmt.Sprint(arr), " ", ", ", -1), "[]")
}

// Docodes a base58 encoded string
func decodeBase58(encoded string) string {
	decoded, _ := base58.Decode(encoded)
	return string(decoded)
}
