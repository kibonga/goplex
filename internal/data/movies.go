package data

import (
	"database/sql"
	"errors"
	"time"

	"github.com/lib/pq"
	"goplex.kibonga/internal/validator"
)

type Movie struct {
	Id        int64     `json:"id"`
	CreatedAt time.Time `json:"-"`
	Title     string    `json:"title"`
	Year      int32     `json:"released,omitempty"`
	Runtime   Runtime   `json:"runtime,omitempty"`
	Genres    []string  `json:"genres,omitempty"`
	Version   int32     `json:"version,omitempty"`
}

type MovieModel struct {
	DB *sql.DB
}

func ValidateMovie(v *validator.Validator, m *Movie) {
	validateTitle(v, m.Title)
	validateYear(v, m.Year)
	validateRuntime(v, &m.Runtime)
	validateGenres(v, m.Genres...)
}

func validateTitle(v *validator.Validator, title string) {
	v.Check(requiredTitle(title), "title", "is required")
	v.Check(maxTitleLen(title), "title", "must not be more than 500 bytes long")
}

func requiredTitle(title string) bool {
	return title != ""
}

func maxTitleLen(title string) bool {
	return len(title) <= 500
}

func validateYear(v *validator.Validator, year int32) {
	v.Check(requiredYear(year), "year", "is required")
	v.Check(minYear(year), "year", "must be greater than 1888")
	v.Check(maxYear(year), "year", "must not be in future")
}

func requiredYear(year int32) bool {
	return year != 0
}

func minYear(year int32) bool {
	return year >= 1888
}

func maxYear(year int32) bool {
	return year <= int32(time.Now().Year())
}

func validateRuntime(v *validator.Validator, runtime *Runtime) {
	v.Check(requiredRuntime(runtime), "runtime", "is required")
	v.Check(nonNegativeRuntime(runtime), "runtime", "must be positive number")
}

func requiredRuntime(runtime *Runtime) bool {
	return *runtime != 0
}

func nonNegativeRuntime(runtime *Runtime) bool {
	return *runtime > 0
}

func validateGenres(v *validator.Validator, genres ...string) {
	v.Check(requiredGenres(genres...), "genres", "must contain at least one genre")
	v.Check(maxGenres(genres), "genres", "must not contain more than 5 genres")
	v.Check(uniqueGenres(genres), "genres", "must not contain duplicates")
}

func requiredGenres(genres ...string) bool {
	return len(genres) > 0
}

func maxGenres(genres []string) bool {
	return len(genres) <= 5
}

func uniqueGenres(genres []string) bool {
	return validator.Unique(genres...)
}

func (m MovieModel) Insert(movie *Movie) error {
	query := `insert into movies (title, year, runtime, genres)
	values ($1, $2, $3, $4)
	returning id, created_at, version`

	args := []interface{}{movie.Title, movie.Year, movie.Runtime, pq.Array(movie.Genres)}

	return m.DB.QueryRow(query, args...).Scan(&movie.Id, &movie.CreatedAt, &movie.Version)
}

func (m MovieModel) Update(movie *Movie) error {
	query := `update movies
	set title = $1, year = $2, runtime = $3, genres = $4, version = version + 1
	where id = $5
	returning version`

	args := []interface{}{movie.Title, movie.Year, movie.Runtime, pq.Array(movie.Genres), movie.Id}

	return m.DB.QueryRow(query, args...).Scan(&movie.Version)
}

func (m MovieModel) Get(id int) (*Movie, error) {
	if id <= 0 {
		return nil, ErrRecordNotFound
	}

	query := `select id, created_at, title, year, runtime, genres, version
	from movies where id = $1`

	movie := Movie{}

	err := m.DB.QueryRow(query, id).Scan(&movie.Id, &movie.CreatedAt, &movie.Title, &movie.Year, &movie.Runtime, pq.Array(&movie.Genres), &movie.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &movie, nil
}

func (m MovieModel) Delete(id int) error {
	return nil
}
