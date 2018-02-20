package main

import (
	"github.com/gin-gonic/gin"

	"bytes"
	"database/sql"
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

func (s server) playsByTimeHandler(c *gin.Context) {
	usernames := parseUsernames(c.Param("usernames"))

	var monthRows *sql.Rows
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
		monthRows, err = s.db.Query(queryBuffer.String())
		defer monthRows.Close()
	} else {
		monthRows, err = s.dotSelect.Query(s.db, "select-plays-by-month")
		defer monthRows.Close()
	}

	if err != nil {
		s.internalServerError(c, err.Error())
		return
	}

	var playsByMonthLabel []string
	var playsByMonthData []int

	for monthRows.Next() {
		var key string
		var count int
		err = monthRows.Scan(&key, &count)
		if err != nil {
			s.internalServerError(c, err.Error())
			return
		}
		playsByMonthLabel = append(playsByMonthLabel, key)
		playsByMonthData = append(playsByMonthData, count)
	}

	var hourRows *sql.Rows
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
		hourRows, err = s.db.Query(queryBuffer.String())
		defer hourRows.Close()
	} else {
		hourRows, err = s.dotSelect.Query(s.db, "select-plays-by-hour")
		defer hourRows.Close()
	}

	if err != nil {
		s.internalServerError(c, err.Error())
		return
	}

	playsByHourData := make([]int, 24)

	for hourRows.Next() {
		var key int
		var count int
		err = hourRows.Scan(&key, &count)
		if err != nil {
			s.internalServerError(c, err.Error())
			return
		}
		playsByHourData[key] = count
	}

	usernameRows, err := s.dotSelect.Query(s.db, "select-usernames")
	defer usernameRows.Close()

	if err != nil {
		s.internalServerError(c, err.Error())
		return
	}

	var allUsernames []string
	for usernameRows.Next() {
		var name string
		err := usernameRows.Scan(&name)
		if err != nil {
			s.internalServerError(c, err.Error())
			return
		}
		allUsernames = append(allUsernames, name)
	}

	c.HTML(http.StatusOK, "playsByTime", gin.H{
		"title":             "Plays By Time",
		"usernames":         allUsernames,
		"playsByTimeTab":    true,
		"playsByMonthLabel": playsByMonthLabel,
		"playsByMonthData":  playsByMonthData,
		"playsByHourData":   playsByHourData,
	})
}
