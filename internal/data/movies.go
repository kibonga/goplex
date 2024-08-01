package data

import "time"

type Movie struct {
	Id        int64     `json:"id"`
	CreatedAt time.Time `json:"-"`
	Title     string    `json:"title"`
	Year      int32     `json:"released,omitempty"`
	Runtime   int32     `json:"length,omitempty"`
	Genres    []string  `json:"genres,omitempty"`
	Version   int32     `json:"version,omitempty"`
}
