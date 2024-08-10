package data

import (
	"database/sql"
	"errors"
)

type Models struct {
	Movies MovieModel
}

var (
	ErrRecordNotFound = errors.New("record not found")
)

func NewModels(db *sql.DB) Models {
	return Models{
		Movies: MovieModel{DB: db},
	}
}
