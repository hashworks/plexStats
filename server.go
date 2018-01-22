package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	// gin/logger.go might report undefined: isatty.IsCygwinTerminal
	// Fix: go get -u github.com/mattn/go-isatty

	"database/sql"
	"flag"
	"github.com/gchaincl/dotsql"
	_ "github.com/mattn/go-sqlite3"
	"os"
)

type server struct {
	db     *sql.DB
	router *gin.Engine
}

func main() {
	flag.Parse()

	var s server
	var err error

	s.db, err = sql.Open("sqlite3", "plex.db")
	defer s.db.Close()
	if err != nil {
		fmt.Printf("Failed to open the database: %s\n", err.Error())
		os.Exit(1)
	}

	// Init database
	dotInit, err := dotsql.LoadFromFile("sql/init.sql")
	for _, command := range []string{
		"create-table-event",
		"create-view-playsByHour",
		"create-trigger-rating",
		"create-table-media",
		"create-table-filter",
		"create-table-server",
		"create-table-account",
		"create-table-client",
		"create-table-address",
		"create-table-hasDirector",
		"create-table-hasProducer",
		"create-table-isSimilarWith",
		"create-table-hasWriter",
		"create-table-hasRole",
		"create-table-hasGenre",
		"create-table-isFromCountry",
		"create-table-isInCollection",
	} {
		_, err := dotInit.Exec(s.db, command)
		if err != nil {
			fmt.Printf("Failed to init database: %s failed with %s\n", command, err.Error())
			os.Exit(1)
		}
	}

	//gin.SetMode(gin.ReleaseMode)
	s.router = gin.Default()

	s.router.StaticFile("/scripts/Chart.min.js", "node_modules/chart.js/dist/Chart.min.js")
	s.router.StaticFile("/", "static/index.html")
	s.router.LoadHTMLGlob("templates/*.html")

	s.router.GET("/playsByHour/*usernames", s.playsByHourHandler)
	s.router.POST("/webhook", s.backendHandler)

	s.router.Run(":8080")
}
