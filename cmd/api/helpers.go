package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/julienschmidt/httprouter"
	"goplex.kibonga/internal/validator"
)

type payload map[string]interface{}

func (app *app) readIdParam(r *http.Request) (int64, error) {
	params := httprouter.ParamsFromContext(r.Context())

	id, err := strconv.ParseInt(params.ByName("id"), 10, 32)
	if err != nil || id < 1 {
		return 0, fmt.Errorf("invalid id provided")
	}

	return id, nil
}

func (app *app) readStr(qs url.Values, k string, def string) string {
	s := qs.Get(k)

	if s == "" {
		return def
	}

	return s
}

func (app *app) readCSV(qs url.Values, k string, def []string) []string {
	csv := qs.Get(k)

	if csv == "" {
		return def
	}

	return strings.Split(csv, ",")
}

func (app *app) readInt(qs url.Values, k string, v *validator.Validator, def int) int {
	s := qs.Get(k)

	if s == "" {
		return def
	}

	i, err := strconv.ParseInt(s, 10, 32)
	if err != nil {
		v.AddError(k, "must be an integer value")
		return def
	}

	return int(i)
}

func (app *app) writeJson(w http.ResponseWriter, status int, data interface{}, headers http.Header) error {
	payload, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		return err
	}

	payload = append(payload, '\n')

	for k, v := range headers {
		w.Header()[k] = v
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(payload)

	return nil
}

func (app *app) writeJsonToStream(w http.ResponseWriter, status int, data interface{}, headers http.Header) error {
	for k, v := range headers {
		for _, vv := range v {
			w.Header().Add(k, vv)
		}
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		return err
	}

	return nil
}

func (app *app) decodeJson(r *http.Request, data interface{}) error {
	err := json.NewDecoder(r.Body).Decode(data)
	defer r.Body.Close()

	if err != nil {
		var syntaxErr *json.SyntaxError
		var unmarshalTypeErr *json.UnmarshalTypeError
		var invalidUnmarshalErr *json.InvalidUnmarshalError

		switch {
		case errors.As(err, &syntaxErr):
			return fmt.Errorf("body contains badly formed JSON (at char %d)", syntaxErr.Offset)
		case errors.As(err, &unmarshalTypeErr):
			if unmarshalTypeErr.Field != "" {
				return fmt.Errorf("body contains incorrect JSON type for field %q", unmarshalTypeErr.Field)
			}
			return fmt.Errorf("body contains incorrect JSON type (at char %d)", unmarshalTypeErr.Offset)
		case errors.As(err, &invalidUnmarshalErr):
			panic(err)
		case errors.Is(err, io.EOF):
			return errors.New("body must not be empty")
		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("body contains badly formed JSON")
		default:
			return err
		}
	}
	r.Body.Close()

	return nil
}

func (app *app) unmarshalJson(r *http.Request, data interface{}) error {
	b, err := io.ReadAll(r.Body)
	if err != nil {
		// This err cannot be caught since io.ReadAll doesn't treat
		// EOF as error
		// if err == io.ErrUnexpectedEOF {
		// 	return errors.New("body contains badly formed JSON")
		// }
		return err
	}
	if len(b) < 1 {
		return errors.New("body must not be empty")
	}

	err = json.Unmarshal(b, data)
	if err != nil {
		var syntaxErr *json.SyntaxError
		var unmarshalTypeErr *json.UnmarshalTypeError
		var invalidUnmarshalErr *json.InvalidUnmarshalError

		switch {
		// This error also cannot be caught since json.Unmarshal treats
		// empty body as Syntax error
		// case errors.Is(err, io.ErrUnexpectedEOF):
		// 	return errors.New("body contains badly formed JSON")
		case errors.As(err, &syntaxErr):
			return fmt.Errorf("body contains badly formed JSON (at char %d)", syntaxErr.Offset)
		case errors.As(err, &unmarshalTypeErr):
			if unmarshalTypeErr.Field != "" {
				return fmt.Errorf("body contains incorrect JSON type for field %q", unmarshalTypeErr.Field)
			}
			return fmt.Errorf("body contains incorrect JSON type (at char %d)", unmarshalTypeErr.Offset)
		case errors.As(err, &invalidUnmarshalErr):
			panic(err)
		default:
			return err
		}
	}

	defer r.Body.Close()
	return nil
}
