package data

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
	"goplex.kibonga/internal/validator"
)

var (
	ErrDuplicateEmail        = errors.New("duplicate email")
	duplicateEmailConstraint = `duplicate key value violates unique constraint "users_email_key"`
)

var AnonymousUser = &User{}

type User struct {
	Id        int64     `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Password  password  `json:"-"`
	Activated bool      `json:"activated"`
	Version   int32     `json:"-"`
}

type UserModel struct {
	DB *sql.DB
}

type password struct {
	plaintext *string
	hash      []byte
}

func (p *password) Set(plaintextPassword string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(plaintextPassword), 12)
	if err != nil {
		return err
	}

	p.hash = hash
	p.plaintext = &plaintextPassword

	return nil
}

func (p *password) Matches(plaintextPassword string) (bool, error) {
	err := bcrypt.CompareHashAndPassword(p.hash, []byte(plaintextPassword))
	if err != nil {
		switch {
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			return false, err
		default:
			return false, nil
		}
	}

	return true, nil
}

func ValidateUser(v *validator.Validator, u *User) {
	validateName(v, u.Name)
	ValidateEmail(v, u.Email)
	ValidatePassword(v, &u.Password)
}

func ValidateEmail(v *validator.Validator, email string) {
	v.Check(v.RequiredString(email), "email", "is required")
	v.Check(validEmail(email), "email", "must be a valid email address")
}

func validEmail(email string) bool {
	return validator.EmailRegExp.MatchString(email)
}

func ValidatePassword(v *validator.Validator, p *password) {
	if p.plaintext != nil {
		ValidatePasswordPlaintext(v, *p.plaintext)
	}

	if p.hash == nil {
		panic("missing password hash for user")
	}
}

func ValidatePasswordPlaintext(v *validator.Validator, password string) {
	v.Check(len(password) > 0, "password", "is required")
	v.Check(minPasswordLen(password), "password", "must be at least 8 bytes long")
	v.Check(maxPasswordLen(password), "password", "must not be more than 72 bytes")
}

func minPasswordLen(password string) bool {
	return len(password) >= 8
}

func maxPasswordLen(password string) bool {
	return len(password) <= 72
}

func validateName(v *validator.Validator, name string) {
	v.Check(v.RequiredString(name), "name", "must be provided")
	v.Check(minNameLen(name), "name", "must be at least 2 bytes")
	v.Check(maxNameLen(name), "name", "must not be more than 500 bytes long")
}

func minNameLen(name string) bool {
	return len(name) >= 2
}

func maxNameLen(name string) bool {
	return len(name) <= 500
}

func (m UserModel) Insert(u *User) error {
	query := `insert into users (name, email, password_hash, activated)
	values($1, $2, $3, $4)
	returning id, created_at, version`

	args := []interface{}{u.Name, u.Email, u.Password.hash, u.Activated}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&u.Id, &u.CreatedAt, &u.Version)
	if err != nil {
		switch {
		case strings.Contains(err.Error(), duplicateEmailConstraint):
			return ErrDuplicateEmail
		default:
			return err
		}
	}

	return nil
}

func (m UserModel) GetByEmail(email string) (*User, error) {
	query := `select id, created_at, name, password_hash, activated, version
	from users
	where email = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var u User

	err := m.DB.QueryRowContext(ctx, query, email).Scan(&u.Id, &u.CreatedAt, &u.Name, &u.Password.hash, &u.Activated, &u.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	u.Email = email

	return &u, nil
}

func (m UserModel) Update(u *User) error {
	query := `update users
	set name = $1, email = $2, password_hash = $3, activated = $4, version = version + 1
	where id = $5 and version = $6
	returning version`

	args := []interface{}{u.Name, u.Email, u.Password.hash, u.Activated, u.Id, u.Version}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&u.Version)
	if err != nil {
		switch {
		case strings.Contains(err.Error(), duplicateEmailConstraint):
			return ErrDuplicateEmail
		case errors.Is(err, sql.ErrNoRows):
			return ErrRecordNotFound
		default:
			return err
		}
	}

	return nil
}

func (m UserModel) GetByToken(tokenScope, token string) (*User, error) {
	query := `select id, created_at, name, email, password_hash, activated, version from users u
	inner join tokens t on u.id = t.user_id
	where t.hash = $1 and t.scope = $2 and t.expiry > $3`

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	token_hash := sha256.Sum256([]byte(token))

	args := []interface{}{token_hash[:], tokenScope, time.Now()}

	var user User

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(
		&user.Id,
		&user.CreatedAt,
		&user.Name,
		&user.Email,
		&user.Password.hash,
		&user.Activated,
		&user.Version,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &user, nil
}

func (u *User) IsAnonymous() bool {
	return u == AnonymousUser
}
