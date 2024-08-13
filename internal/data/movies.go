package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
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

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	return m.DB.QueryRowContext(ctx, query, args...).Scan(&movie.Id, &movie.CreatedAt, &movie.Version)
}

func (m MovieModel) Update(movie *Movie) error {
	query := `update movies
	set title = $1, year = $2, runtime = $3, genres = $4, version = version + 1
	where id = $5 and version = $6
	returning version`

	args := []interface{}{movie.Title, movie.Year, movie.Runtime, pq.Array(movie.Genres), movie.Id, movie.Version}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&movie.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}

	return nil
}

func (m MovieModel) Get(id int) (*Movie, error) {
	if id <= 0 {
		return nil, ErrRecordNotFound
	}

	query := `select id, created_at, title, year, runtime, genres, version
	from movies where id = $1`

	movie := Movie{}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, id).Scan(&movie.Id, &movie.CreatedAt, &movie.Title, &movie.Year, &movie.Runtime, &movie.Genres, &movie.Version)
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
	query := `delete from movies
	where id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	sqlRes, err := m.DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := sqlRes.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil
}

func (m MovieModel) GetAll(title string, genres []string, filters *Filters) ([]*Movie, Metadata, error) {
	query := fmt.Sprintf(`select count(*) over(), id, created_at, title, year, runtime, genres, version
	from movies 
	where (to_tsvector('simple', title) @@ plainto_tsquery('simple', $1) or $1 = '') and
	(genres @> $2 or $2 = '{}')
	order by %s %s, id asc limit $3 offset $4`, filters.sortColumn(), filters.sortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	args := []interface{}{title, pq.Array(genres), filters.limit(), filters.offset()}

	sqlRows, err := m.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, Metadata{}, err
	}
	defer sqlRows.Close()

	totalRecords := 0
	movies := []*Movie{}

	for sqlRows.Next() {
		var m Movie

		err = sqlRows.Scan(&totalRecords, &m.Id, &m.CreatedAt, &m.Title, &m.Year, &m.Runtime, pq.Array(&m.Genres), &m.Version)
		if err != nil {
			return nil, Metadata{}, err
		}

		movies = append(movies, &m)
	}

	if err := sqlRows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)
	return movies, metadata, nil
}
