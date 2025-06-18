package main

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

// struct tags that rename field i.e. Error to error
type jsonResponse struct {
	Error   bool   `json:"error"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

func (app *Config) readJSON(w http.ResponseWriter, r *http.Request, data any) error {
	maxBytes := 1048576 // one mb

	// limits size of incoming body
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	// parse JSON
	dec := json.NewDecoder(r.Body)
	err := dec.Decode(data)

	if err != nil {
		return err
	}

	// check only single JSON value
	// decode tries to read another JSON value, with parameter being dummy place to dump any data
	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		return errors.New("body must have only a single JSON value")
	}

	return nil
}

// ..., status code, value, headers
// with below jsonResponse
// out is created, no headers (would just merge them)
// header line tells client to interpret as JSON
// sends 404 status to client
// sets JSON response to response body
// sends response to client of error, header, then body
func (app *Config) writeJSON(w http.ResponseWriter, status int, data any, headers ...http.Header) error {
	out, err := json.Marshal(data)
	if err != nil {
		return err
	}

	if len(headers) > 0 {
		for key, value := range headers[0] {
			w.Header()[key] = value
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, err = w.Write(out)
	if err != nil {
		return err
	}

	return nil
}

// ... = 0 or more errors
// passing in w, errors.New("item not found"), StatusNotFound
// builds
//
//	jsonResponse {
//		Error: true,
//		Message: "item not found",
//	 Data: nil,
//	}
func (app *Config) errorJSON(w http.ResponseWriter, err error, status ...int) error {
	// sets default error as 400
	statusCode := http.StatusBadRequest

	// uses first status code
	if len(status) > 0 {
		statusCode = status[0]
	}

	// struct with error, message, data
	var payload jsonResponse
	payload.Error = true
	payload.Message = err.Error()

	return app.writeJSON(w, statusCode, payload)
}
