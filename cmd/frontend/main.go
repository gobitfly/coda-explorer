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
	"coda-explorer/handlers"
	"coda-explorer/services"
	"coda-explorer/util"
	"flag"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	negronilogrus "github.com/meatballhat/negroni-logrus"
	"github.com/phyber/negroni-gzip/gzip"
	"github.com/urfave/negroni"
	"github.com/zesik/proxyaddr"
)

var logger = logrus.New().WithField("module", "main")

func main() {
	dbHost := flag.String("dbHost", "", "Database host")
	dbPort := flag.String("dbPort", "", "Database port")
	dbUser := flag.String("dbUser", "", "Database user")
	dbPassword := flag.String("dbPassword", "", "Database password")
	dbName := flag.String("dbName", "", "Database name")

	port := flag.Int("port", 3333, "Port to start the frontend http server on")

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

	router := mux.NewRouter()
	router.HandleFunc("/", handlers.Index).Methods("GET")
	router.HandleFunc("/index/data", handlers.IndexPageData).Methods("GET")
	router.HandleFunc("/block/{hash}", handlers.Block).Methods("GET")
	router.HandleFunc("/blocks", handlers.Blocks).Methods("GET")
	router.HandleFunc("/blocks/data", handlers.BlocksData).Methods("GET")
	router.HandleFunc("/status", handlers.Status).Methods("GET")

	router.PathPrefix("/").Handler(http.FileServer(http.Dir("static")))

	n := negroni.New(negroni.NewRecovery())

	// Customize the logging middleware to include a proper module entry for the frontend
	frontendLogger := negronilogrus.NewMiddleware()
	frontendLogger.Before = func(entry *logrus.Entry, request *http.Request, s string) *logrus.Entry {
		entry = negronilogrus.DefaultBefore(entry, request, s)
		return entry.WithField("module", "frontend")
	}
	frontendLogger.After = func(entry *logrus.Entry, writer negroni.ResponseWriter, duration time.Duration, s string) *logrus.Entry {
		entry = negronilogrus.DefaultAfter(entry, writer, duration, s)
		return entry.WithField("module", "frontend")
	}
	n.Use(frontendLogger)

	n.Use(gzip.Gzip(gzip.DefaultCompression))

	pa := &proxyaddr.ProxyAddr{}
	pa.Init(proxyaddr.CIDRLoopback)
	n.Use(pa)

	n.UseHandler(router)

	services.Init()

	srv := &http.Server{
		Addr:         fmt.Sprintf("0.0.0.0:%v", *port),
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      n,
	}

	log.Printf("http server listening on %v", srv.Addr)
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Println(err)
		}
	}()

	util.WaitForCtrlC()
}
