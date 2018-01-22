package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	// gin/logger.go might report undefined: isatty.IsCygwinTerminal
	// Fix: go get -u github.com/mattn/go-isatty

	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"os"
)

type server struct {
	db     *sql.DB
	router *gin.Engine
}

func main() {
	var s server
	var err error

	s.db, err = sql.Open("sqlite3", "data/plex.db")
	defer s.db.Close()
	if err != nil {
		fmt.Printf("Failed to open the database: %s\n", err.Error())
		os.Exit(1)
	}

	//gin.SetMode(gin.ReleaseMode)
	s.router = gin.Default()

	s.router.StaticFile("/scripts/Chart.min.js", "node_modules/chart.js/dist/Chart.min.js")
	//s.router.LoadHTMLGlob("templates/*")
	s.router.LoadHTMLFiles("templates/index.tmpl")

	s.router.GET("/", s.frontendHandler)
	s.router.POST("/webhook", s.backendHandler)

	s.router.Run(":8080")
}
