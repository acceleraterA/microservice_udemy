package main

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

type jsonResponse struct {
	Error   bool        `json:"error"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// read json

func (app *Config) readJSON(w http.ResponseWriter, r *http.Request, data any) error {
	maxBytes := 1048576                                      //limitation of the size of the uploaded json file, one Megabyte
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes)) //read the json file
	dec := json.NewDecoder(r.Body)                           //decode the json file
	err := dec.Decode(data)                                  //	decode the json file
	if err != nil {
		return err
	}
	// check if only on json file received
	err = dec.Decode(&struct{}{}) //check if the json file is empty
	if err != io.EOF {
		return errors.New("request body must only have a single JSON object")
	}
	return nil
}

// write json
func (app *Config) writeJSON(w http.ResponseWriter, status int, data interface{}, headers ...http.Header) error {
	out, err := json.Marshal(data)
	if err != nil {
		return err
	}
	if len(headers) > 0 {
		for k, v := range headers[0] {
			w.Header()[k] = v
		}
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, err = w.Write(out)
	if err != nil {
		return err
	}
	return err
}

//write error msg as json

func (app *Config) errorJSON(w http.ResponseWriter, err error, status ...int) error {
	statusCode := http.StatusBadRequest
	if len(status) > 0 {
		statusCode = status[0]
	}
	payload := jsonResponse{
		Error:   true,
		Message: err.Error(),
	}
	return app.writeJSON(w, statusCode, payload)
}
