package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	// gin/logger.go might report undefined: isatty.IsCygwinTerminal
	// Fix: go get -u github.com/mattn/go-isatty
	"net/http"
	"time"

	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"os"
)

func main() {
	db, err := sql.Open("sqlite3", "data/plex.db")
	defer db.Close()
	if err != nil {
		fmt.Printf("Failed to open the database: %s\n", err.Error())
		os.Exit(1)
	}

	//gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	/**
	 * Webfrontent
	 */

	router.StaticFile("/scripts/Chart.min.js", "node_modules/chart.js/dist/Chart.min.js")
	//router.LoadHTMLGlob("templates/*")
	router.LoadHTMLFiles("templates/index.tmpl")
	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.tmpl", gin.H{
			"title": "Foo",
			"foo":   "Bar",
		})
	})

	/**
	 * Webhook
	 */

	router.POST("/webhook", func(c *gin.Context) {
		// Add `date` parameter to parse previously logged requests
		date := c.Query("date") // Always string
		t, err := time.Parse(time.RFC3339, date)
		if err != nil {
			// Empty or false parameter? Current time
			t = time.Now()
		}

		// Parse JSON
		var event Event
		if err := c.ShouldBindJSON(&event); err == nil {
			// Begin transaction
			if tx, err := db.Begin(); err == nil {
				// Create rollback function in case shit goes downhill
				rollback := func(error string) {
					// TODO: Better error logging
					tx.Rollback()
					fmt.Println(error)
					c.Status(http.StatusInternalServerError)
				}

				/**
				 **** TABLES ****
				 */

				/**
				 * Type statement
				 */
				// Preparation
				eventStmt, err := tx.Prepare("INSERT INTO event(date, type, rating, local, owned) VALUES(?, ?, ?, ?, ?)")
				defer eventStmt.Close()
				if err != nil {
					rollback(err.Error())
					return
				}

				var eventType string
				if event.IsMediaPlay() {
					eventType = "play"
				} else if event.IsMediaPause() {
					eventType = "pause"
				} else if event.IsMediaResume() {
					eventType = "resume"
				} else if event.IsMediaStop() {
					eventType = "stop"
				} else if event.IsMediaRating() {
					eventType = "userRating"
				} else if event.IsMediaScrobble() {
					eventType = "scrobble"
				} else {
					rollback(fmt.Sprintf("Unknown event type '%s'", event.Type))
					return
				}

				// Execution
				eventResult, err := eventStmt.Exec(t.Format(time.RFC3339), eventType, event.Rating, event.Player.Local, event.Owned)
				if err != nil {
					rollback(err.Error())
					return
				}
				// Get primary key
				eventId, err := eventResult.LastInsertId()
				if err != nil {
					rollback(err.Error())
					return
				}

				/**
				 * Server statement
				 */
				// Try to update
				serverUpdateStmt, err := tx.Prepare("UPDATE OR FAIL server SET `name` = ? WHERE uuid = ?")
				defer serverUpdateStmt.Close()
				if err != nil {
					rollback(err.Error())
					return
				}
				serverUpdateResult, err := serverUpdateStmt.Exec(event.Server.Name, event.Server.UUID)
				if err != nil {
					rollback(err.Error())
					return
				}
				if serverUpdateRowCount, err := serverUpdateResult.RowsAffected(); err != nil || serverUpdateRowCount == 0 {
					// Insert new server
					serverStmt, err := tx.Prepare("INSERT INTO server(uuid, `name`) VALUES(?, ?)")
					defer serverStmt.Close()
					if err != nil {
						rollback(err.Error())
						return
					}
					_, err = serverStmt.Exec(event.Server.UUID, event.Server.Name)
					if err != nil {
						rollback(err.Error())
						return
					}
				}

				/**
				 * Account statement
				 */
				// Try to update
				accountUpdateStmt, err := tx.Prepare("UPDATE OR FAIL account SET `name`= ?, thumbnail = ? WHERE plexnumber = ?")
				defer accountUpdateStmt.Close()
				if err != nil {
					rollback(err.Error())
					return
				}
				accountUpdateResult, err := accountUpdateStmt.Exec(event.Account.Name, event.Account.Thumb, event.Account.ID)
				if err != nil {
					rollback(err.Error())
					return
				}
				if accountUpdateRowCount, err := accountUpdateResult.RowsAffected(); err != nil || accountUpdateRowCount == 0 {
					// Insert new account
					accountStmt, err := tx.Prepare("INSERT INTO account(plexNumber, `name`, thumbnail) VALUES(?, ?, ?)")
					defer accountStmt.Close()
					if err != nil {
						rollback(err.Error())
						return
					}
					_, err = accountStmt.Exec(event.Account.ID, event.Account.Name, event.Account.Thumb)
					if err != nil {
						rollback(err.Error())
						return
					}
				}

				/**
				 * Address statement
				 */
				// Check if IP exists already
				var addressId int64
				lastAddressQuery, err := tx.Prepare("SELECT aId FROM address WHERE ip = ? LIMIT 1")
				defer lastAddressQuery.Close()
				if err != nil {
					rollback(err.Error())
					return
				}
				err = lastAddressQuery.QueryRow(event.Player.Address).Scan(&addressId)
				if err != nil {
					// Insert otherwise
					addressStmt, err := tx.Prepare("INSERT INTO address(ip) VALUES(?)")
					defer addressStmt.Close()
					if err != nil {
						rollback(err.Error())
						return
					}
					addressResult, err := addressStmt.Exec(event.Player.Address)
					if err != nil {
						rollback(err.Error())
						return
					}
					addressId, err = addressResult.LastInsertId()
					if err != nil {
						rollback(err.Error())
						return
					}
				}

				/**
				 * Client statement
				 */
				// Try to update
				clientUpdateStmt, err := tx.Prepare("UPDATE OR FAIL client SET `name`= ?, lastAddressId = ? WHERE uuid = ?")
				defer clientUpdateStmt.Close()
				if err != nil {
					rollback(err.Error())
					return
				}
				clientUpdateResult, err := clientUpdateStmt.Exec(event.Player.Name, addressId, event.Player.UUID)
				if err != nil {
					rollback(err.Error())
					return
				}
				if clientUpdateRowCount, err := clientUpdateResult.RowsAffected(); err != nil || clientUpdateRowCount == 0 {
					// Insert new client
					clientStmt, err := tx.Prepare("INSERT INTO client(uuid, `name`, lastAddressId) VALUES(?, ?, ?)")
					defer clientStmt.Close()
					if err != nil {
						rollback(err.Error())
						return
					}
					_, err = clientStmt.Exec(event.Player.UUID, event.Player.Name, addressId)
					if err != nil {
						rollback(err.Error())
						return
					}
				}

				/**
				 * media statement
				 */
				// Preparation
				var mediaType string
				if event.Metadata.IsMovie() {
					mediaType = "movie"
				} else if event.Metadata.IsEpisode() {
					mediaType = "episode"
				} else if event.Metadata.IsTrack() {
					mediaType = "track"
				} else if event.Metadata.IsImage() {
					mediaType = "image"
				} else if event.Metadata.IsClip() {
					mediaType = "clip"
				} else {
					rollback(fmt.Sprintf("Unknown media type '%s'", event.Metadata.Type))
					return
				}

				var subType string
				if event.Metadata.IsTrailer() {
					subType = "trailer"
				} else if event.Metadata.SubType != "" {
					rollback(fmt.Sprintf("Unknown media subtype '%s'", event.Metadata.SubType))
					return
				}

				var dateOriginalRFC3339 string
				dateOriginal, err := event.Metadata.OriginallyAvailableAt()
				if err != nil {
					dateOriginalRFC3339 = ""
				} else {
					dateOriginalRFC3339 = dateOriginal.Format(time.RFC3339)
				}

				// Try to update
				mediaUpdateStmt, err := tx.Prepare("UPDATE OR FAIL media SET " +
					"title = ?," +
					"desc = ?," +
					"type = ?," +
					"subtype = ?," +
					"contentRating = ?," +
					"webRating = ?," +
					"thumbnail = ?," +
					"art = ?," +
					"releaseYear = ?," +
					"dateOriginal = ?," +
					"dateAdded = ?," +
					"dateUpdated = ?" +
					" WHERE guid = ?")
				defer mediaUpdateStmt.Close()
				if err != nil {
					rollback(err.Error())
					return
				}

				mediaUpdateResult, err := mediaUpdateStmt.Exec(
					event.Metadata.Title, event.Metadata.Summary, mediaType, subType,
					event.Metadata.ContentRating, event.Metadata.WebRating, event.Metadata.Thumb, event.Metadata.Art,
					event.Metadata.ReleaseYear, dateOriginalRFC3339, event.Metadata.AddedAt().Format(time.RFC3339),
					event.Metadata.UpdatedAt().Format(time.RFC3339), event.Metadata.GUID)
				if err != nil {
					rollback(err.Error())
					return
				}

				if mediaUpdateRowCount, err := mediaUpdateResult.RowsAffected(); err != nil || mediaUpdateRowCount == 0 {
					// Insert new event
					mediaStmt, err := tx.Prepare("INSERT INTO media(" +
						"guid, title, `desc`, type, subtype, contentRating, webRating, thumbnail, art, releaseYear, " +
						"dateOriginal, dateAdded, dateUpdated" +
						") VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
					defer mediaStmt.Close()
					if err != nil {
						rollback(err.Error())
						return
					}
					_, err = mediaStmt.Exec(event.Metadata.GUID,
						event.Metadata.Title, event.Metadata.Summary, mediaType, subType,
						event.Metadata.ContentRating, event.Metadata.WebRating, event.Metadata.Thumb, event.Metadata.Art,
						event.Metadata.ReleaseYear, dateOriginalRFC3339, event.Metadata.AddedAt().Format(time.RFC3339),
						event.Metadata.UpdatedAt().Format(time.RFC3339))
					if err != nil {
						rollback(err.Error())
						return
					}
				}

				/**
				 **** RELATIONS ****
				 */

				/**
				 * triggeredEvent statement
				 */
				triggeredEventStmt, err := tx.Prepare("INSERT INTO triggeredEvent(plexNumber, uuid, eId) VALUES(?, ?, ?)")
				defer triggeredEventStmt.Close()
				if err != nil {
					rollback(err.Error())
					return
				}
				_, err = triggeredEventStmt.Exec(event.Account.ID, event.Server.UUID, eventId)
				if err != nil {
					rollback(err.Error())
					return
				}

				/**
				 * usedMedia statement
				 */
				usedMediaStmt, err := tx.Prepare("INSERT INTO usedMedia(eId, guid) VALUES(?, ?)")
				defer usedMediaStmt.Close()
				if err != nil {
					rollback(err.Error())
					return
				}
				_, err = usedMediaStmt.Exec(eventId, event.Metadata.GUID)
				if err != nil {
					rollback(err.Error())
					return
				}

				/**
				 * usedClient statement
				 */
				usedClientStmt, err := tx.Prepare("INSERT INTO usedClient(eId, uuid) VALUES(?, ?)")
				defer usedClientStmt.Close()
				if err != nil {
					rollback(err.Error())
					return
				}
				_, err = usedClientStmt.Exec(eventId, event.Player.UUID)
				if err != nil {
					rollback(err.Error())
					return
				}

				/**
				 * clientUsedAddress statement
				 */
				clientUsedAddressStmt, err := tx.Prepare("INSERT OR IGNORE INTO clientUsedAddress(uuid, aId) VALUES(?, ?)")
				defer clientUsedAddressStmt.Close()
				if err != nil {
					rollback(err.Error())
					return
				}
				_, err = clientUsedAddressStmt.Exec(event.Player.UUID, addressId)
				if err != nil {
					rollback(err.Error())
					return
				}

				// Commit the transaction
				tx.Commit()
				c.Status(http.StatusOK)
			} else {
				// Failed to create the transaction, stop
				// TODO: Better error logging
				fmt.Println(err.Error())
				c.Status(http.StatusInternalServerError)
				return
			}
		} else {
			// Failed to parse the JSON, stop
			// TODO: Better error logging
			fmt.Println(err.Error())
			c.Status(http.StatusBadRequest)
			return
		}
	})

	router.Run(":8080")
}
