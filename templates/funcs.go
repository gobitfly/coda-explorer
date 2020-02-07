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
	"html/template"
	"time"

	"github.com/leekchan/gtf"
)

func GetTemplateFuncs() template.FuncMap {
	fm := template.FuncMap{
		"formatSeconds":      formatSeconds,
		"formatMilliSeconds": formatMilliSeconds,
	}

	gtf.ForceInject(fm)
	return fm
}

func formatSeconds(seconds int) string {
	return fmt.Sprintf("%v", time.Second*time.Duration(seconds))
}

func formatMilliSeconds(ms int) string {
	return fmt.Sprintf("%v", time.Millisecond*time.Duration(ms))
}
