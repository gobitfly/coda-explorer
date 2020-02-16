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

package util

import (
	"log"
	"os"
	"os/signal"
	"runtime/debug"
	"strconv"
	"time"
)

// MustParseInt must parse an int
func MustParseInt(str string) int {
	res, err := strconv.Atoi(str)

	if err != nil {
		debug.PrintStack()
		log.Fatalf("error parsing string %v: %v", str, err)
	}

	return res
}

// MustParseJsTimestamp must parse a timestamp in js timestamp format
func MustParseJsTimestamp(str string) time.Time {
	res, err := strconv.ParseInt(str, 10, 64)

	if err != nil {
		log.Fatalf("error parsing string %v: %v", str, err)
	}

	ts := time.Unix(res/1000, 0)

	return ts
}

// WaitForCtrlC will block/wait until a control-c is pressed
func WaitForCtrlC() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
}
