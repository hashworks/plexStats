package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	// gin/logger.go might report undefined: isatty.IsCygwinTerminal
	// Fix: go get -u github.com/mattn/go-isatty
	"github.com/elazarl/go-bindata-assetfs"
	"github.com/gin-contrib/multitemplate"

	"database/sql"
	"github.com/gchaincl/dotsql"
	_ "github.com/mattn/go-sqlite3"
	"html/template"
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
	dotInit, err := dotsql.LoadFromString(string(MustAsset("sql/init.sql")))
	if err != nil {
		panic(err)
	}

	// Load alter commands
	s.dotAlter, err = dotsql.LoadFromString(string(MustAsset("sql/alter.sql")))
	if err != nil {
		panic(err)
	}

	// Load select commands
	s.dotSelect, err = dotsql.LoadFromString(string(MustAsset("sql/select.sql")))
	if err != nil {
		panic(err)
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

	//s.router.SetFuncMap(templateFunctionMap())

	// Load template file names from Asset
	templateNames, err := AssetDir("templates")
	if err != nil {
		panic(err)
	}

	// Create a base template where we add the template functions
	tmpl := template.New("")
	tmpl.Funcs(templateFunctionMap())

	// Iterate trough template files, load them into multitemplate
	multiT := multitemplate.New()
	for _, templateName := range templateNames {
		basename := templateName[:len(templateName)-5]
		tmpl := tmpl.New(basename)
		tmpl, err := tmpl.Parse(string(MustAsset("templates/" + templateName)))
		if err != nil {
			panic(err)
		}
		multiT.Add(basename, tmpl)
		fmt.Printf("Loaded templates/%s as %s\n", templateName, templateName[:len(templateName)-5])
	}
	// multitemplate is our new HTML renderer
	s.router.HTMLRender = multiT

	s.router.StaticFS("/scripts/", &assetfs.AssetFS{Asset: Asset, AssetDir: AssetDir, AssetInfo: AssetInfo, Prefix: "node_modules/chart.js/dist"})
	s.router.StaticFS("/css/", &assetfs.AssetFS{Asset: Asset, AssetDir: AssetDir, AssetInfo: AssetInfo, Prefix: "css"})

	s.router.Any("/", s.indexHandler)
	s.router.Any("/playsByTime/*usernames", s.playsByTimeHandler)
	s.router.POST("/webhook", s.backendHandler)

	s.router.Run(":8080")
}
