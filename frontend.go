package main

import (
	"github.com/gin-gonic/gin"

	"bytes"
	"fmt"
	"net/http"
	"strings"
)

func (s server) playsByHourHandler(c *gin.Context) {
	internalServerError := func(err string) {
		// TODO: Better error logging
		fmt.Println(err)
		c.Status(http.StatusInternalServerError)
	}
	usernamesParam := strings.Trim(c.Param("usernames"), "/")
	var usernames []string
	if usernamesParam != "" {
		usernames = strings.Split(usernamesParam, "/")
	}

	var queryBuffer bytes.Buffer
	if len(usernames) > 0 {
		// Get plays by hour of a person.
		// Since date is a RFC3339 string we need to… do stuff.
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
	} else {
		queryBuffer.WriteString("SELECT * FROM playsByHour")
	}

	rows, err := s.db.Query(queryBuffer.String())
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
		"playsByHourData": playsByHour,
	})
}
