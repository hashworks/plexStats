package main

import (
	"flag"
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

var (
	// Set the following uppercase three with -ldflags "-X main.VERSION=v1.2.3 [...]"
	VERSION      string = "unknown"
	BUILD_COMMIT string = "unknown"
	BUILD_DATE   string = "unknown"
	versionFlag  bool
	address      string
	port         int
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
	flagSet := flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	flagSet.BoolVar(&versionFlag, "version", false, "Displays the version and license information.")
	flagSet.StringVar(&address, "address", "127.0.0.1", "The address to listen on.")
	flagSet.IntVar(&port, "port", 65431, "The port to listen on.")

	flagSet.Parse(os.Args[1:])

	switch {
	case versionFlag:
		fmt.Println("plexStats")
		fmt.Println("https://github.com/hashworks/plexStats")
		fmt.Println("Version: " + VERSION)
		fmt.Println("Commit: " + BUILD_COMMIT)
		fmt.Println("Build date: " + BUILD_DATE)
		fmt.Println()
		fmt.Println("Published under the GNU General Public License v3.0.")
	default:
		run()
	}
}

func run() {

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

	s.router.GET("/", s.indexHandler)
	s.router.GET("/playsByTime/*usernames", s.playsByTimeHandler)
	s.router.POST("/webhook", s.backendHandler)

	s.router.Run(fmt.Sprintf("%s:%d", address, port))
}
