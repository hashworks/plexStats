package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func (s server) indexHandler(c *gin.Context) {
	var (
		eventCount   int
		accountCount int
		clientCount  int
		movieCount   int
		episodeCount int
		trackCount   int
	)

	rows, err := s.dotSelect.Query(s.db, "count-events")
	if err != nil {
		s.internalServerError(c, err.Error())
		return
	}
	rows.Next()
	err = rows.Scan(&eventCount)
	if err != nil {
		s.internalServerError(c, err.Error())
		return
	}

	rows, err = s.dotSelect.Query(s.db, "count-accounts")
	if err != nil {
		s.internalServerError(c, err.Error())
		return
	}
	rows.Next()
	err = rows.Scan(&accountCount)
	if err != nil {
		s.internalServerError(c, err.Error())
		return
	}

	rows, err = s.dotSelect.Query(s.db, "count-clients")
	if err != nil {
		s.internalServerError(c, err.Error())
		return
	}
	rows.Next()
	err = rows.Scan(&clientCount)
	if err != nil {
		s.internalServerError(c, err.Error())
		return
	}

	rows, err = s.dotSelect.Query(s.db, "count-movies")
	if err != nil {
		s.internalServerError(c, err.Error())
		return
	}
	rows.Next()
	err = rows.Scan(&movieCount)
	if err != nil {
		s.internalServerError(c, err.Error())
		return
	}

	rows, err = s.dotSelect.Query(s.db, "count-episodes")
	if err != nil {
		s.internalServerError(c, err.Error())
		return
	}
	rows.Next()
	err = rows.Scan(&episodeCount)
	if err != nil {
		s.internalServerError(c, err.Error())
		return
	}

	rows, err = s.dotSelect.Query(s.db, "count-tracks")
	if err != nil {
		s.internalServerError(c, err.Error())
		return
	}
	rows.Next()
	err = rows.Scan(&trackCount)
	if err != nil {
		s.internalServerError(c, err.Error())
		return
	}

	c.HTML(http.StatusOK, "index", gin.H{
		"title":        "",
		"eventCount":   eventCount,
		"accountCount": accountCount,
		"clientCount":  clientCount,
		"movieCount":   movieCount,
		"episodeCount": episodeCount,
		"trackCount":   trackCount,
	})
}
