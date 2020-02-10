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

package main

import (
	"coda-explorer/db"
	"flag"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	"log"
	"time"
)

var logger = logrus.New().WithField("module", "main")

// Helper application to re-generate all statistics
func main() {
	dbHost := flag.String("dbHost", "", "Database host")
	dbPort := flag.String("dbPort", "", "Database port")
	dbUser := flag.String("dbUser", "", "Database user")
	dbPassword := flag.String("dbPassword", "", "Database password")
	dbName := flag.String("dbName", "", "Database name")

	flag.Parse()

	dbConn, err := sqlx.Open("postgres", fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", *dbUser, *dbPassword, *dbHost, *dbPort, *dbName))
	if err != nil {
		logger.Fatal(err)
	}
	// The golang postgres sql driver does not properly implement PingContext
	// therefore we use a timer to catch db connection timeouts
	dbConnectionTimeout := time.NewTimer(15 * time.Second)
	go func() {
		<-dbConnectionTimeout.C
		log.Fatal("Timeout while connecting to the database")
	}()
	err = dbConn.Ping()
	if err != nil {
		logger.Fatal(err)
	}
	dbConnectionTimeout.Stop()

	logger.Info("database connection established")

	db.DB = dbConn
	defer db.DB.Close()

	var startTime time.Time
	err = db.DB.Get(&startTime, "SELECT MIN(ts) FROM blocks")
	if err != nil {
		logger.Fatalf("error retrieving start time from blocks table: %v", err)
	}

	currTime := startTime.Truncate(time.Hour * 24)
	endTime := time.Now().Truncate(time.Hour * 24)
	for currTime.Before(endTime) {
		logger.Infof("exporting statistics for day %v", currTime)
		err := db.GenerateAndSaveStatistics(currTime)
		if err != nil {
			logger.Fatalf("error generating statistics for day %v: %v", currTime, err)
		}
		currTime = currTime.Add(time.Hour * 24)
	}
	logger.Infof("regenerated all statistics")
}
