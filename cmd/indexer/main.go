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
	"coda-explorer/indexer"
	"coda-explorer/util"
	"flag"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	"log"
	"time"

	_ "github.com/lib/pq"
)

var logger = logrus.New().WithField("module", "main")

func main() {
	dbHost := flag.String("dbHost", "", "Database host")
	dbPort := flag.String("dbPort", "", "Database port")
	dbUser := flag.String("dbUser", "", "Database user")
	dbPassword := flag.String("dbPassword", "", "Database password")
	dbName := flag.String("dbName", "", "Database name")

	codaEndpoint := flag.String("coda", "localhost:3085/graphql", "CODA node graphql endpoint")
	startupLookback := flag.Int("startupLookback", 1000, "Check the last x blocks immediately after startup")

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

	indexer.Start(*codaEndpoint, *startupLookback)

	util.WaitForCtrlC()

}
