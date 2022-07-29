package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

func (s server) backendHandler(c *gin.Context) {
	// Add `date` parameter to parse previously logged requests
	date := c.Query("date") // Always string
	t, err := time.Parse(time.RFC3339, date)
	if err != nil {
		// Empty or false parameter? Current time
		t = time.Now()
	}

	jsonData, exists := c.GetPostForm("payload")
	if !exists {
		fmt.Println("No payload found.")
		c.Status(http.StatusBadRequest)
		return
	}

	// Parse JSON
	var event Event
	if err := json.Unmarshal([]byte(jsonData), &event); err == nil {
		// Begin transaction
		if tx, err := s.db.Begin(); err == nil {
			// Create rollback function in case shit goes downhill
			rollback := func(where string, error string) {
				// TODO: Better error logging
				tx.Rollback()
				fmt.Printf("Rollback at '%s' with error '%s'\n", where, error)
				c.Status(http.StatusInternalServerError)
			}

			/**
			 **** TABLES ****
			 */

			/**
			 * Server statement
			 */
			// Try to update
			serverUpdateStmt, err := s.dotAlter.Prepare(tx, "update-server")
			defer serverUpdateStmt.Close()
			if err != nil {
				rollback("update-server prepare", err.Error())
				return
			}
			serverUpdateResult, err := serverUpdateStmt.Exec(event.Server.Name, event.Server.UUID)
			if err != nil {
				rollback("update-server exec", err.Error())
				return
			}
			if serverUpdateRowCount, err := serverUpdateResult.RowsAffected(); err != nil || serverUpdateRowCount == 0 {
				// Insert new server
				serverStmt, err := s.dotAlter.Prepare(tx, "insert-server")
				defer serverStmt.Close()
				if err != nil {
					rollback("insert-server prepare", err.Error())
					return
				}
				_, err = serverStmt.Exec(event.Server.UUID, event.Server.Name)
				if err != nil {
					rollback("insert-server exec", err.Error())
					return
				}
			}

			/**
			 * Account statement
			 */
			// Try to update
			accountUpdateStmt, err := s.dotAlter.Prepare(tx, "update-account")
			defer accountUpdateStmt.Close()
			if err != nil {
				rollback("update-account prepare", err.Error())
				return
			}
			accountUpdateResult, err := accountUpdateStmt.Exec(event.Account.Name, event.Account.Thumb, event.Account.ID)
			if err != nil {
				rollback("update-account exec", err.Error())
				return
			}
			if accountUpdateRowCount, err := accountUpdateResult.RowsAffected(); err != nil || accountUpdateRowCount == 0 {
				// Insert new account
				accountStmt, err := s.dotAlter.Prepare(tx, "insert-account")
				defer accountStmt.Close()
				if err != nil {
					rollback("insert-account prepare", err.Error())
					return
				}
				_, err = accountStmt.Exec(event.Account.ID, event.Account.Name, event.Account.Thumb)
				if err != nil {
					rollback("insert-account exec", err.Error())
					return
				}
			}

			/**
			 * Address statement
			 */
			// Check if IP exists already
			var addressId int64
			lastAddressQuery, err := s.dotAlter.Prepare(tx, "select-address-id-by-ip")
			defer lastAddressQuery.Close()
			if err != nil {
				rollback("select-address-id-by-ip prepare", err.Error())
				return
			}
			err = lastAddressQuery.QueryRow(event.Player.Address).Scan(&addressId)
			if err != nil {
				// Insert otherwise
				addressStmt, err := s.dotAlter.Prepare(tx, "insert-address")
				defer addressStmt.Close()
				if err != nil {
					rollback("insert-address prepare", err.Error())
					return
				}
				addressResult, err := addressStmt.Exec(event.Player.Address)
				if err != nil {
					rollback("insert-address exec", err.Error())
					return
				}
				addressId, err = addressResult.LastInsertId()
				if err != nil {
					rollback("addressId", err.Error())
					return
				}
			}

			/**
			 * Client statement
			 */
			// Try to update
			clientUpdateStmt, err := s.dotAlter.Prepare(tx, "update-client")
			defer clientUpdateStmt.Close()
			if err != nil {
				rollback("update-client prepare", err.Error())
				return
			}
			clientUpdateResult, err := clientUpdateStmt.Exec(event.Player.Name, event.Player.UUID)
			if err != nil {
				rollback("update-client exec", err.Error())
				return
			}
			if clientUpdateRowCount, err := clientUpdateResult.RowsAffected(); err != nil || clientUpdateRowCount == 0 {
				// Insert new client
				clientStmt, err := s.dotAlter.Prepare(tx, "insert-client")
				defer clientStmt.Close()
				if err != nil {
					rollback("insert-client prepare", err.Error())
					return
				}
				_, err = clientStmt.Exec(event.Player.UUID, event.Player.Name)
				if err != nil {
					rollback("insert-client exec", err.Error())
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
				rollback("mediaType", fmt.Sprintf("Unknown media type '%s'", event.Metadata.Type))
				return
			}

			var subType string
			if event.Metadata.IsTrailer() {
				subType = "trailer"
			} else if event.Metadata.SubType != "" {
				rollback("subtype", fmt.Sprintf("Unknown media subtype '%s'", event.Metadata.SubType))
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
			mediaUpdateStmt, err := s.dotAlter.Prepare(tx, "update-media")
			defer mediaUpdateStmt.Close()
			if err != nil {
				rollback("update-media prepare", err.Error())
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
				rollback("update-media exec", err.Error())
				return
			}

			if mediaUpdateRowCount, err := mediaUpdateResult.RowsAffected(); err != nil || mediaUpdateRowCount == 0 {
				// Insert new event
				mediaStmt, err := s.dotAlter.Prepare(tx, "insert-media")
				defer mediaStmt.Close()
				if err != nil {
					rollback("insert-media prepare", err.Error())
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
					rollback("insert-media exec", err.Error())
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
				rollback("eventType", fmt.Sprintf("Unknown event type '%s'", event.Type))
				return
			}

			eventStmt, err := s.dotAlter.Prepare(tx, "insert-event")
			defer eventStmt.Close()
			if err != nil {
				rollback("insert-event prepare", err.Error())
				return
			}

			// Execution
			_, err = eventStmt.Exec(t.Format(time.RFC3339), eventType, event.Rating, event.Player.Local,
				event.Owned, event.Account.ID, event.Server.UUID, event.Player.UUID, event.Metadata.GUID, addressId)
			if err != nil {
				rollback("insert-event exec", err.Error())
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
						filterUpdateStmt, err := s.dotAlter.Prepare(tx, "update-filter")
						defer filterUpdateStmt.Close()
						if err != nil {
							rollback("update-filter prepare", err.Error())
							return
						}
						filterUpdateResult, err := filterUpdateStmt.Exec(filter.Tag, filter.Filter, filter.Role, filter.Thumb, filter.Count, filter.Id)
						if err != nil {
							rollback("update-filter exec", err.Error())
							return
						}
						if filterUpdateRowCount, err := filterUpdateResult.RowsAffected(); err != nil || filterUpdateRowCount == 0 {
							// Insert new filter
							filterStmt, err := s.dotAlter.Prepare(tx, "insert-filter")
							defer filterStmt.Close()
							if err != nil {
								rollback("insert-filter prepare", err.Error())
								return
							}
							_, err = filterStmt.Exec(filter.Id, filter.Tag, filter.Filter, filter.Role, filter.Thumb, filter.Count)
							if err != nil {
								rollback("insert-filter exec", err.Error())
								return
							}
						}
						// Add relation
						relationStmt, err := tx.Prepare(fmt.Sprintf("INSERT OR IGNORE INTO %s(guid, fId) VALUES(?, ?)", relationTable))
						defer relationStmt.Close()
						if err != nil {
							rollback("add-relation prepare", err.Error())
							return
						}
						_, err = relationStmt.Exec(event.Metadata.GUID, filter.Id)
						if err != nil {
							rollback("add-relation exec", err.Error())
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
			fmt.Printf("Failed to create the transaction: %s\n", err.Error())
			c.Status(http.StatusInternalServerError)
			return
		}
	} else {
		// Failed to parse the JSON, stop
		// TODO: Better error logging
		fmt.Printf("Failed to parse JSON: %s\n", err.Error())

		// Plex changes their JSON format quite often, so we need a way to log this
		tmpDir, err := ioutil.TempDir("", "")
		if err == nil {
			file, err := os.Create(tmpDir + string(os.PathSeparator) + "invalid_plex_webhook_body.json")
			if err == nil {
				defer file.Close()
				file.WriteString(jsonData)
			}
		}

		c.Status(http.StatusBadRequest)
		return
	}
}
