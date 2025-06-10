package main

import (
	"fmt"
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

func (cfg *apiConfig) counterHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	//w.Write([]byte("Hits: " + fmt.Sprint(cfg.fileserverHits.Load())))
	fmt.Fprintf(w, "<html><body><h1>Welcome, Chirpy Admin</h1><p>Chirpy has been visited %d times!</p></body></html>", cfg.fileserverHits.Load())
}

func (cfg *apiConfig) resetHandler(w http.ResponseWriter, req *http.Request) {
	cfg.fileserverHits.Store(0)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("counter reset"))
}

func readinessHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(http.StatusText(http.StatusOK)))
}

func validationHandler(w http.ResponseWriter, req *http.Request) {
	reqData := unmarshalJson(req)

	if len(reqData.Body) <= 140 {
		newBody := wordFilter(reqData.Body)
		respondWithJSON(w, http.StatusOK, jsonCleaned{CleanedBody: newBody})
	} else if len(reqData.Body) > 140 {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long")
	} else {
		respondWithError(w, http.StatusBadRequest, "Something went wrong")
	}
	/*data := marshalJson(resData)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write(data)*/
}
