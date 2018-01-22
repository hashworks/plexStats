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
					os.Exit(1) // TODO: Remove after debug
				}

				/**
				 **** TABLES ****
				 */

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
				clientUpdateStmt, err := tx.Prepare("UPDATE OR FAIL client SET `name`= ? WHERE uuid = ?")
				defer clientUpdateStmt.Close()
				if err != nil {
					rollback(err.Error())
					return
				}
				clientUpdateResult, err := clientUpdateStmt.Exec(event.Player.Name, event.Player.UUID)
				if err != nil {
					rollback(err.Error())
					return
				}
				if clientUpdateRowCount, err := clientUpdateResult.RowsAffected(); err != nil || clientUpdateRowCount == 0 {
					// Insert new client
					clientStmt, err := tx.Prepare("INSERT INTO client(uuid, `name`) VALUES(?, ?)")
					defer clientStmt.Close()
					if err != nil {
						rollback(err.Error())
						return
					}
					_, err = clientStmt.Exec(event.Player.UUID, event.Player.Name)
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
					"type = ?, " +
					"subtype = ?, " +

					"key = ?, " +
					"parentKey = ?, " +
					"grandparentKey = ?, " +
					"primaryExtraKey = ?, " +

					"title = ?, " +
					"titleSort = ?, " +
					"parentTitle = ?, " +
					"grandparentTitle = ?, " +

					"summary = ?, " +
					"duration = ?, " +

					"thumb = ?, " +
					"parentThumb = ?, " +
					"grandparentThumb = ?, " +

					"grandparentTheme = ?, " +
					"grandparentRatingKey = ?, " +

					"art = ?, " +
					"grandparentArt = ?, " +

					"`index` = ?, " +
					"parentIndex = ?, " +

					"studio = ?, " +
					"tagline = ?, " +
					"chapterSource = ?, " +

					"librarySectionID = ?, " +
					"librarySectionKey = ?, " +
					"librarySectionType = ?, " +

					"webRating = ?, " +
					"userRating = ?, " +
					"audienceRating = ?, " +
					"contentRating = ?, " +
					"ratingImage = ?, " +
					"viewCount = ?, " +

					"releaseYear = ?, " +
					"dateOriginal = ?, " +
					"dateAdded = ?, " +
					"dateUpdated = ? " +
					" WHERE guid = ?")
				defer mediaUpdateStmt.Close()
				if err != nil {
					rollback(err.Error())
					return
				}

				mediaUpdateResult, err := mediaUpdateStmt.Exec(
					mediaType,
					subType,

					event.Metadata.Key,
					event.Metadata.ParentKey,
					event.Metadata.GrandparentKey,
					event.Metadata.PrimaryExtraKey,

					event.Metadata.Title,
					event.Metadata.TitleSort,
					event.Metadata.ParentTitle,
					event.Metadata.GrandparentTitle,

					event.Metadata.Summary,
					event.Metadata.Duration,

					event.Metadata.Thumb,
					event.Metadata.ParentThumb,
					event.Metadata.GrandparentThumb,

					event.Metadata.GrandparentTheme,
					event.Metadata.GrandparentRatingKey,

					event.Metadata.Art,
					event.Metadata.GrandparentArt,

					event.Metadata.Index,
					event.Metadata.ParentIndex,

					event.Metadata.Studio,
					event.Metadata.Tagline,
					event.Metadata.ChapterSource,

					event.Metadata.LibrarySectionID,
					event.Metadata.LibrarySectionKey,
					event.Metadata.LibrarySectionType,

					event.Metadata.WebRating,
					event.Metadata.UserRating,
					event.Metadata.AudienceRating,
					event.Metadata.ContentRating,
					event.Metadata.RatingImage,
					event.Metadata.ViewCount,

					event.Metadata.ReleaseYear,
					dateOriginalRFC3339,
					event.Metadata.AddedAt().Format(time.RFC3339),
					event.Metadata.UpdatedAt().Format(time.RFC3339),

					event.Metadata.GUID)
				if err != nil {
					rollback(err.Error())
					return
				}

				if mediaUpdateRowCount, err := mediaUpdateResult.RowsAffected(); err != nil || mediaUpdateRowCount == 0 {
					// Insert new event
					mediaStmt, err := tx.Prepare("INSERT INTO media(" +
						"guid, " +

						"type, " +
						"subtype, " +

						"key, " +
						"parentKey, " +
						"grandparentKey, " +
						"primaryExtraKey, " +

						"title, " +
						"titleSort, " +
						"parentTitle, " +
						"grandparentTitle, " +

						"summary, " +
						"duration, " +

						"thumb, " +
						"parentThumb, " +
						"grandparentThumb, " +

						"grandparentTheme, " +
						"grandparentRatingKey, " +

						"art, " +
						"grandparentArt, " +

						"`index`, " +
						"parentIndex, " +

						"studio, " +
						"tagline, " +
						"chapterSource, " +

						"librarySectionID, " +
						"librarySectionKey, " +
						"librarySectionType, " +

						"webRating, " +
						"userRating, " +
						"audienceRating, " +
						"contentRating, " +
						"ratingImage, " +
						"viewCount, " +

						"releaseYear, " +
						"dateOriginal, " +
						"dateAdded, " +
						"dateUpdated" +
						") VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
					defer mediaStmt.Close()
					if err != nil {
						rollback(err.Error())
						return
					}
					_, err = mediaStmt.Exec(
						event.Metadata.GUID,

						mediaType,
						subType,

						event.Metadata.Key,
						event.Metadata.ParentKey,
						event.Metadata.GrandparentKey,
						event.Metadata.PrimaryExtraKey,

						event.Metadata.Title,
						event.Metadata.TitleSort,
						event.Metadata.ParentTitle,
						event.Metadata.GrandparentTitle,

						event.Metadata.Summary,
						event.Metadata.Duration,

						event.Metadata.Thumb,
						event.Metadata.ParentThumb,
						event.Metadata.GrandparentThumb,

						event.Metadata.GrandparentTheme,
						event.Metadata.GrandparentRatingKey,

						event.Metadata.Art,
						event.Metadata.GrandparentArt,

						event.Metadata.Index,
						event.Metadata.ParentIndex,

						event.Metadata.Studio,
						event.Metadata.Tagline,
						event.Metadata.ChapterSource,

						event.Metadata.LibrarySectionID,
						event.Metadata.LibrarySectionKey,
						event.Metadata.LibrarySectionType,

						event.Metadata.WebRating,
						event.Metadata.UserRating,
						event.Metadata.AudienceRating,
						event.Metadata.ContentRating,
						event.Metadata.RatingImage,
						event.Metadata.ViewCount,

						event.Metadata.ReleaseYear,
						dateOriginalRFC3339,
						event.Metadata.AddedAt().Format(time.RFC3339),
						event.Metadata.UpdatedAt().Format(time.RFC3339))
					if err != nil {
						rollback(err.Error())
						return
					}
				}

				/**
				 * Event statement
				 */
				// Preparation
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

				eventStmt, err := tx.Prepare("INSERT INTO event(date, type, rating, local, owned," +
					"accountNumber, sUUID, cUUID, mGUID, aId) " +
					"VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
				defer eventStmt.Close()
				if err != nil {
					rollback(err.Error())
					return
				}

				// Execution
				_, err = eventStmt.Exec(t.Format(time.RFC3339), eventType, event.Rating, event.Player.Local,
					event.Owned, event.Account.ID, event.Server.UUID, event.Player.UUID, event.Metadata.GUID, addressId)
				if err != nil {
					rollback(err.Error())
					return
				}

				/**
				 * filter statements and relations
				 */
				filterRelations := map[string][]Filter{
					"hasDirector":    event.Metadata.Director,
					"hasProducer":    event.Metadata.Producer,
					"isSimilarWith":  event.Metadata.Similar,
					"hasWriter":      event.Metadata.Writer,
					"hasRole":        event.Metadata.Role,
					"hasGenre":       event.Metadata.Genre,
					"isFromCountry":  event.Metadata.Country,
					"isInCollection": event.Metadata.Collection,
				}
				// Walk trough filter relations
				for relationTable, filters := range filterRelations {
					for _, filter := range filters {
						/*
						 * Since we use defer inside a for loop we need to encapsulate this
						 * https://blog.zkanda.io/defer-inside-a-for-loop/
						 */
						func() {
							// Try to update
							filterUpdateStmt, err := tx.Prepare("UPDATE OR FAIL filter SET " +
								"tag = ?, filter = ?, role = ?, thumb = ?, count = ? WHERE fId = ?")
							defer filterUpdateStmt.Close()
							if err != nil {
								rollback(err.Error())
								return
							}
							filterUpdateResult, err := filterUpdateStmt.Exec(filter.Tag, filter.Filter, filter.Role, filter.Thumb, filter.Count, filter.Id)
							if err != nil {
								rollback(err.Error())
								return
							}
							if filterUpdateRowCount, err := filterUpdateResult.RowsAffected(); err != nil || filterUpdateRowCount == 0 {
								// Insert new filter
								filterStmt, err := tx.Prepare("INSERT INTO filter(fId, tag, filter, role, thumb, count) " +
									"VALUES(?, ?, ?, ?, ?, ?)")
								defer filterStmt.Close()
								if err != nil {
									rollback(err.Error())
									return
								}
								_, err = filterStmt.Exec(filter.Id, filter.Tag, filter.Filter, filter.Role, filter.Thumb, filter.Count)
								if err != nil {
									rollback(err.Error())
									return
								}
							}
							// Add relation
							relationStmt, err := tx.Prepare(fmt.Sprintf("INSERT OR IGNORE INTO %s(guid, fId) VALUES(?, ?)", relationTable))
							defer relationStmt.Close()
							if err != nil {
								rollback(err.Error())
								return
							}
							_, err = relationStmt.Exec(event.Metadata.GUID, filter.Id)
							if err != nil {
								rollback(err.Error())
								return
							}
						}()
					}
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
