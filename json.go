package main

import (
	"encoding/json"
	"log"
	"net/http"
)

type jsonError struct {
	JsonError string `json:"error"`
}

func marshalJson(res interface{}) []byte {
	data, err := json.Marshal(res)
	if err != nil {
		log.Printf("Unable to marshal the request: %s ", res)
	}
	return data
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	data := marshalJson(payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(data)
}

func respondWithError(w http.ResponseWriter, code int, msg string) {
	respondWithJSON(w, code, jsonError{JsonError: msg})
}
