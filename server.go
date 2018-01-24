package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	// gin/logger.go might report undefined: isatty.IsCygwinTerminal
	// Fix: go get -u github.com/mattn/go-isatty

	"database/sql"
	"github.com/gchaincl/dotsql"
	_ "github.com/mattn/go-sqlite3"
	"net/http"
	"os"
)

type server struct {
	dotAlter  *dotsql.DotSql
	dotSelect *dotsql.DotSql
	db        *sql.DB
	router    *gin.Engine
}

func (s server) internalServerError(c *gin.Context, err string) {
	// TODO: Better error logging
	fmt.Println(err)
	c.Status(http.StatusInternalServerError)
}

func main() {
	var s server
	var err error

	// Load init commands
	dotInit, err := dotsql.LoadFromFile("sql/init.sql")
	if err != nil {
		fmt.Printf("Failed to open the database init file: %s\n", err.Error())
		os.Exit(1)
	}

	// Load alter commands
	s.dotAlter, err = dotsql.LoadFromFile("sql/alter.sql")
	if err != nil {
		fmt.Printf("Failed to open the database alter command file: %s\n", err.Error())
		os.Exit(1)
	}

	// Load select commands
	s.dotSelect, err = dotsql.LoadFromFile("sql/select.sql")
	if err != nil {
		fmt.Printf("Failed to open the database select command file: %s\n", err.Error())
		os.Exit(1)
	}

	// Open database file
	s.db, err = sql.Open("sqlite3", "plex.db")
	defer s.db.Close()
	if err != nil {
		fmt.Printf("Failed to open the database: %s\n", err.Error())
		os.Exit(1)
	}

	// Init database
	for _, command := range []string{
		"create-table-event",
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
	s.router.StaticFile("/css/main.css", "css/main.css")
	s.router.Static("/fonts/", "fonts/")
	s.router.LoadHTMLGlob("templates/*.html")

	s.router.GET("/", s.indexHandler)
	s.router.GET("/playsByTime/*usernames", s.playsByTimeHandler)
	s.router.POST("/webhook", s.backendHandler)

	s.router.Run(":8080")
}
