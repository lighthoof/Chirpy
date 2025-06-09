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

type jsonValid struct {
	JsonValid bool `json:"valid"`
}

/*func marshalJson(res any) {
	switch value := res.(type) {
	case jsonError:

	}
	data, err := json.Marshal(res)
}*/

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
