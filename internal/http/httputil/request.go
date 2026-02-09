package httputil

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

const DefaultMaxJSONBodyBytes int64 = 1 << 20 // 1 MiB

var (
	ErrInvalidJSON  = errors.New("invalid json")
	ErrBodyTooLarge = errors.New("request body too large")
)

// DecodeJSON decodes JSON body into dst using strict rules:
// - limited body size
// - unknown fields are rejected
// - only a single JSON object is allowed in body
func DecodeJSON(w http.ResponseWriter, r *http.Request, dst any, maxBytes int64) error {
	if maxBytes > 0 {
		r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
	}

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(dst); err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			return ErrBodyTooLarge
		}
		return fmt.Errorf("%w: %v", ErrInvalidJSON, err)
	}

	if err := dec.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return ErrInvalidJSON
	}

	return nil
}

// DecodeJSONOrError decodes request JSON and writes a normalized 400 error response on failure.
func DecodeJSONOrError(w http.ResponseWriter, r *http.Request, dst any, maxBytes int64) bool {
	if err := DecodeJSON(w, r, dst, maxBytes); err != nil {
		if errors.Is(err, ErrBodyTooLarge) {
			Error(w, http.StatusBadRequest, "request body too large")
			return false
		}
		Error(w, http.StatusBadRequest, "invalid json")
		return false
	}
	return true
}
