package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
)

type jsonBody struct {
	Body string `json:"body"`
}

type jsonError struct {
	JsonError string `json:"error"`
}

type jsonCleaned struct {
	CleanedBody string `json:"cleaned_body"`
}

func marshalJson(res interface{}) []byte {
	data, err := json.Marshal(res)
	if err != nil {
		log.Printf("Unable to marshal the request: %s ", res)
	}
	return data
}

func unmarshalJson(req *http.Request) jsonBody {
	reqBody := jsonBody{Body: ""}

	data, err := io.ReadAll(req.Body)
	if err != nil {
		log.Printf("Unable to read the request body: %s %s", req.Method, req.URL.Path)
	}

	err = json.Unmarshal(data, &reqBody)
	if err != nil {
		log.Printf("Unable to unmarshal the request: %s %s", req.Method, req.URL.Path)
	}

	return reqBody
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
