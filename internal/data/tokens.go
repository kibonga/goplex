package data

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base32"
	"time"

	"goplex.kibonga/internal/validator"
)

const (
	ScopeActivation     = "activation"
	ScopeAuthentication = "authentication"
)

type Token struct {
	PlainText string    `json:"token"`
	Hash      []byte    `json:"-"`
	UserID    int64     `json:"-"`
	Expiry    time.Time `json:"expiry"`
	Scope     string    `json:"-"`
}

type TokenModel struct {
	DB *sql.DB
}

func generateToken(userID int64, ttl time.Duration, scope string) (*Token, error) {
	token := &Token{
		UserID: userID,
		Expiry: time.Now().Add(ttl),
		Scope:  scope,
	}

	randBytes := make([]byte, 16)

	_, err := rand.Read(randBytes)
	if err != nil {
		return nil, err
	}

	token.PlainText = base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(randBytes)

	hash := sha256.Sum256([]byte(token.PlainText))
	token.Hash = hash[:]

	return token, nil
}

func ValidateToken(v *validator.Validator, t *Token) {
	ValidatePlaintextToken(v, t.PlainText)
}

func ValidatePlaintextToken(v *validator.Validator, token string) {
	v.Check(token != "", "token", "must be provided")
	v.Check(len(token) == 26, "token", "must be 26 bytes long")
}

func (m TokenModel) New(userID int64, ttl time.Duration, scope string) (*Token, error) {
	token, err := generateToken(userID, ttl, scope)
	if err != nil {
		return nil, err
	}

	err = m.Insert(token)
	if err != nil {
		return nil, err
	}

	return token, nil
}

func (m TokenModel) Insert(t *Token) error {
	query := `insert into tokens (hash, user_id, expiry, scope)
	values ($1, $2, $3, $4)`

	args := []interface{}{t.Hash, t.UserID, t.Expiry, t.Scope}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*300)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, query, args...)
	return err
}

func (m TokenModel) DeleteTokensForUser(scope string, userID int64) error {
	query := `delete from tokens 
	where user_id = $1 and scope = $2`

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, query, userID, scope)
	return err
}
