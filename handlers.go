package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync/atomic"

	"github.com/google/uuid"
	"github.com/lighthoof/Chirpy/internal/database"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	dbQueries      *database.Queries
	platform       string
}

func (cfg *apiConfig) counterHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	//w.Write([]byte("Hits: " + fmt.Sprint(cfg.fileserverHits.Load())))
	fmt.Fprintf(w, "<html><body><h1>Welcome, Chirpy Admin</h1><p>Chirpy has been visited %d times!</p></body></html>", cfg.fileserverHits.Load())
}

func (cfg *apiConfig) resetHandler(w http.ResponseWriter, req *http.Request) {
	cfg.fileserverHits.Store(0)
	if cfg.platform != "dev" {
		respondWithError(w, http.StatusForbidden, "")
	}

	cfg.dbQueries.ClearUsers(req.Context())
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Users cleared"))
}

func readinessHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(http.StatusText(http.StatusOK)))
}

func (cfg *apiConfig) createUserHandler(w http.ResponseWriter, req *http.Request) {
	reqBody := struct {
		Email string `json:"email"`
	}{}

	data, err := io.ReadAll(req.Body)
	if err != nil {
		log.Printf("Unable to read the request body: %s %s", req.Method, req.URL.Path)
		return
	}
	err = json.Unmarshal(data, &reqBody)
	if err != nil {
		log.Printf("Unable to unmarshal the request: %s %s", req.Method, req.URL.Path)
		return
	}

	if len(strings.Split(reqBody.Email, "@")) < 2 {
		log.Printf("Invalid e-mail: %s", reqBody.Email)
		return
	}

	usrDb, err := cfg.dbQueries.CreateUser(req.Context(), reqBody.Email)
	if err != nil {
		log.Printf("Unable to create user with the e-mail: %s %s [%s]", req.Method, req.URL.Path, reqBody.Email)
		return
	}

	user := User{
		ID:        usrDb.ID,
		CreatedAt: usrDb.CreatedAt,
		UpdatedAt: usrDb.UpdatedAt,
		Email:     usrDb.Email,
	}

	respondWithJSON(w, http.StatusCreated, user)
}

func (cfg *apiConfig) createChirpHandler(w http.ResponseWriter, req *http.Request) {
	reqBody := Chirp{}

	data, err := io.ReadAll(req.Body)
	if err != nil {
		log.Printf("Unable to read the request body: %s %s", req.Method, req.URL.Path)
		return
	}

	err = json.Unmarshal(data, &reqBody)
	if err != nil {
		log.Printf("Unable to unmarshal the request: %s %s", req.Method, req.URL.Path)
		return
	}

	if len(reqBody.Body) <= 140 {
		reqBody.Body = wordFilter(reqBody.Body)

		chirpDb, err := cfg.dbQueries.CreateChirp(req.Context(),
			database.CreateChirpParams{Body: reqBody.Body, UserID: reqBody.UserID})
		if err != nil {
			log.Printf("Unable to create chirp: %s %s [%s]", req.Method, req.URL.Path, err)
			return
		}

		respBody := Chirp{
			ID:        chirpDb.ID,
			CreatedAt: chirpDb.CreatedAt,
			UpdatedAt: chirpDb.UpdatedAt,
			Body:      chirpDb.Body,
			UserID:    chirpDb.UserID,
		}
		respondWithJSON(w, http.StatusCreated, respBody)

	} else if len(reqBody.Body) > 140 {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long")
	} else {
		respondWithError(w, http.StatusBadRequest, "Something went wrong")
	}
}

func (cfg *apiConfig) getChirpsHandler(w http.ResponseWriter, req *http.Request) {
	respBody := []Chirp{}

	chirpsDb, err := cfg.dbQueries.GetChirps(req.Context())
	if err != nil {
		log.Printf("Unable to retrieve chirps")
		return
	}
	for _, chirpDb := range chirpsDb {
		chirp := Chirp{
			ID:        chirpDb.ID,
			CreatedAt: chirpDb.CreatedAt,
			UpdatedAt: chirpDb.UpdatedAt,
			Body:      chirpDb.Body,
			UserID:    chirpDb.UserID,
		}
		respBody = append(respBody, chirp)
	}

	respondWithJSON(w, http.StatusOK, respBody)
}

func (cfg *apiConfig) getChirpByIdHandler(w http.ResponseWriter, req *http.Request) {
	chirpID, err := uuid.Parse(req.PathValue("chirpID"))
	if err != nil {
		log.Printf("Unable to parse chirpID: %s", req.PathValue("chirpID"))
		return
	}

	chirpDb, err := cfg.dbQueries.GetChirpById(req.Context(), chirpID)
	if err != nil {
		log.Printf("Unable to retrieve chirps")
		return
	}

	respBody := Chirp{
		ID:        chirpDb.ID,
		CreatedAt: chirpDb.CreatedAt,
		UpdatedAt: chirpDb.UpdatedAt,
		Body:      chirpDb.Body,
		UserID:    chirpDb.UserID,
	}

	respondWithJSON(w, http.StatusOK, respBody)
}
