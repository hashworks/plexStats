package main

import (
	"github.com/gin-gonic/gin"

	"bytes"
	"database/sql"
	"fmt"
	"net/http"
	"strings"
)

func parseUsernames(param string) []string {
	usernamesParam := strings.Trim(param, "/")
	var usernames []string
	if usernamesParam != "" {
		usernames = strings.Split(usernamesParam, "/")
	}
	return usernames
}

func (s server) playsByHourHandler(c *gin.Context) {
	internalServerError := func(err string) {
		// TODO: Better error logging
		fmt.Println(err)
		c.Status(http.StatusInternalServerError)
	}

	usernames := parseUsernames(c.Param("usernames"))

	var rows *sql.Rows
	var err error
	if len(usernames) > 0 {
		// Get plays by hour of a person.
		// Since date is a RFC3339 string we need to… do stuff.
		var queryBuffer bytes.Buffer
		queryBuffer.WriteString(
			"SELECT " +
				"CAST(REPLACE(ROUND(REPLACE(SUBSTR(e.date,12,5),':','.')+0.2), 24, 0) AS INTEGER) AS hour, " +
				"COUNT(e.eId) as count " +
				"FROM event as e, account as a " +
				"WHERE e.type = 'play' AND e.accountNumber = a.plexNumber AND (")
		for i := 0; i < len(usernames); i++ {
			if i > 0 {
				queryBuffer.WriteString(" OR ")
			}
			queryBuffer.WriteString("a.name = '")
			queryBuffer.WriteString(usernames[i])
			queryBuffer.WriteString("'")
		}
		queryBuffer.WriteString(") GROUP BY hour")
		rows, err = s.db.Query(queryBuffer.String())
	} else {
		rows, err = s.dotSelect.Query(s.db, "select-plays-by-hour")
	}

	if err != nil {
		internalServerError(err.Error())
		return
	}

	playsByHour := make([]int, 24)

	for rows.Next() {
		var key int
		var count int
		err = rows.Scan(&key, &count)
		if err != nil {
			internalServerError(err.Error())
			return
		}
		playsByHour[key] = count
	}

	c.HTML(http.StatusOK, "playsByHour.html", gin.H{
		"title":           "Plays By Hour",
		"playsByHourData": playsByHour,
	})
}

func (s server) playsByMonthHandler(c *gin.Context) {
	internalServerError := func(err string) {
		// TODO: Better error logging
		fmt.Println(err)
		c.Status(http.StatusInternalServerError)
	}

	usernames := parseUsernames(c.Param("usernames"))

	var rows *sql.Rows
	var err error
	if len(usernames) > 0 {
		// Get plays by hour of a person.
		// Since date is a RFC3339 string we need to… do stuff.
		var queryBuffer bytes.Buffer
		queryBuffer.WriteString("SELECT SUBSTR(date,0,8) as month, COUNT(eId) as count " +
			"FROM event as e, account as a " +
			"WHERE e.type = 'play' AND e.accountNumber = a.plexNumber AND (")
		for i := 0; i < len(usernames); i++ {
			if i > 0 {
				queryBuffer.WriteString(" OR ")
			}
			queryBuffer.WriteString("a.name = '")
			queryBuffer.WriteString(usernames[i])
			queryBuffer.WriteString("'")
		}
		queryBuffer.WriteString(") GROUP BY month ORDER BY month ASC")
		rows, err = s.db.Query(queryBuffer.String())
	} else {
		rows, err = s.dotSelect.Query(s.db, "select-plays-by-month")
	}

	if err != nil {
		internalServerError(err.Error())
		return
	}

	var playsByMonthLabel []string
	var playsByMonthData []int

	for rows.Next() {
		var key string
		var count int
		err = rows.Scan(&key, &count)
		if err != nil {
			internalServerError(err.Error())
			return
		}
		playsByMonthLabel = append(playsByMonthLabel, key)
		playsByMonthData = append(playsByMonthData, count)
	}

	c.HTML(http.StatusOK, "playsByMonth", gin.H{
		"title":             "Plays By Month",
		"playsByMonthLabel": playsByMonthLabel,
		"playsByMonthData":  playsByMonthData,
	})
}
