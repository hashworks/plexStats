package main

import "time"

type Event struct {
	User     bool     `json:"user"`
	Server   Server   `json:"Server" binding:"required"`
	Type     string   `json:"event" binding:"required"`
	Account  Account  `json:"Account" binding:"required"`
	Owned    bool     `json:"owner"`
	Player   Player   `json:"Player" binding:"required"`
	Metadata Metadata `json:"Metadata" binding:"required"`
	Rating   int      `json:"rating"` // only on IsMediaRating()
}

func (e Event) IsMediaPlay() bool {
	if e.Type == "media.play" {
		return true
	}
	return false
}

func (e Event) IsMediaPause() bool {
	if e.Type == "media.pause" {
		return true
	}
	return false
}

func (e Event) IsMediaResume() bool {
	if e.Type == "media.resume" {
		return true
	}
	return false
}

func (e Event) IsMediaStop() bool {
	if e.Type == "media.stop" {
		return true
	}
	return false
}

func (e Event) IsMediaScrobble() bool {
	if e.Type == "media.scrobble" {
		return true
	}
	return false
}

func (e Event) IsMediaRating() bool {
	if e.Type == "media.rate" {
		return true
	}
	return false
}

type Server struct {
	UUID string `json:"uuid" binding:"required"`
	Name string `json:"title" binding:"required"`
}

type Account struct {
	ID    int    `json:"id" binding:"required"`
	Name  string `json:"title" binding:"required"`
	Thumb string `json:"thumb" binding:"required"`
}

type Player struct {
	UUID    string `json:"uuid" binding:"required"`
	Address string `json:"publicAddress" binding:"required"`
	Name    string `json:"title"`
	Local   bool   `json:"local"`
}

type Metadata struct {
	GUID string `json:"guid" binding:"required"`

	Type    string `json:"type" binding:"required"`
	SubType string `json:"subtype"`

	Key             string `json:"key" binding:"required"`
	ParentKey       string `json:"parentKey"`
	GrandparentKey  string `json:"grandparentKey"`
	PrimaryExtraKey string `json:"primaryExtraKey"`

	Title            string `json:"title" binding:"required"`
	TitleSort        string `json:"titleSort"`
	ParentTitle      string `json:"parentTitle"`
	GrandparentTitle string `json:"grandparentTitle"`

	Summary  string `json:"summary"`
	Duration int    `json:"duration"`

	Thumb            string `json:"thumb"`
	ParentThumb      string `json:"parentThumb"`
	GrandparentThumb string `json:"grandparentThumb"`

	GrandparentTheme     string `json:"grandparentTheme"`
	GrandparentRatingKey int    `json:"grandparentRatingKey,string"`

	Art            string `json:"art"`
	GrandparentArt string `json:"grandparentArt"`

	Index       int `json:"index"`
	ParentIndex int `json:"parentIndex"`

	Studio        string `json:"studio"`
	Tagline       string `json:"tagline"`
	ChapterSource string `json:"chapterSource"`

	LibrarySectionID   int    `json:"librarySectionID"`
	LibrarySectionKey  string `json:"librarySectionKey"`
	LibrarySectionType string `json:"librarySectionType"`

	WebRating      float32 `json:"rating"`
	UserRating     float32 `json:"userRating"`
	AudienceRating float32 `json:"audienceRating"`
	ContentRating  string  `json:"contentRating"`
	RatingImage    string  `json:"ratingImage"`
	ViewCount      int     `json:"viewCount"`

	ReleaseYear                     int    `json:"year"`
	OriginallyAvailableAtDateString string `json:"originallyAvailableAt"`
	AddedAtTimestamp                int64  `json:"addedAt"`
	UpdatedAtTimestamp              int64  `json:"updatedAt"`

	Director   []Filter `json:"director"`
	Producer   []Filter `json:"Producer"`
	Writer     []Filter `json:"Writer"`
	Role       []Filter `json:"Role"`
	Similar    []Filter `json:"Similar"`
	Genre      []Filter `json:"Genre"`
	Country    []Filter `json:"Country"`
	Collection []Filter `json:"Collection"`
}

func (m Metadata) IsEpisode() bool {
	if m.Type == "episode" {
		return true
	}
	return false
}

func (m Metadata) IsMovie() bool {
	if m.Type == "movie" {
		return true
	}
	return false
}

func (m Metadata) IsTrack() bool {
	if m.Type == "track" {
		return true
	}
	return false
}

func (m Metadata) IsClip() bool {
	if m.Type == "clip" {
		return true
	}
	return false
}

func (m Metadata) IsTrailer() bool {
	if m.SubType == "trailer" {
		return true
	}
	return false
}

func (m Metadata) IsImage() bool {
	if m.Type == "image" {
		return true
	}
	return false
}

func (m Metadata) UpdatedAt() time.Time {
	return time.Unix(m.UpdatedAtTimestamp, 0)
}

func (m Metadata) OriginallyAvailableAt() (time.Time, error) {
	return time.Parse("2006-01-02", m.OriginallyAvailableAtDateString)
}

func (m Metadata) AddedAt() time.Time {
	return time.Unix(m.AddedAtTimestamp, 0)
}

type Filter struct {
	Id     int    `json:"id" binding:"required"`
	Tag    string `json:"tag" binding:"required"`
	Filter string `json:"filter" binding:"required"`
	Role   string `json:"role"`
	Thumb  string `json:"thumb"`
	Count  int    `json:"count"`
}
