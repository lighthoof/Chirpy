package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/lighthoof/Chirpy/internal/auth"
	"github.com/lighthoof/Chirpy/internal/database"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	dbQueries      *database.Queries
	platform       string
	secret         string
	authExpiry     time.Duration
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
	reqBody := Auth{}

	var err error = nil
	_ = unmarshalType(req, &reqBody)

	if len(strings.Split(reqBody.Email, "@")) < 2 {
		log.Printf("Invalid e-mail: %s", reqBody.Email)
		return
	}

	reqBody.Password, err = auth.HashPassword(reqBody.Password)
	if err != nil {
		log.Printf("Unable to hash the password: %s", err)
		return
	}

	userDb, err := cfg.dbQueries.CreateUser(req.Context(),
		database.CreateUserParams{Email: reqBody.Email, HashedPassword: reqBody.Password})
	if err != nil {
		log.Printf("Unable to create user with the e-mail: %s %s [%s]", req.Method, req.URL.Path, reqBody.Email)
		return
	}

	user := User{
		ID:          userDb.ID,
		CreatedAt:   userDb.CreatedAt,
		UpdatedAt:   userDb.UpdatedAt,
		Email:       userDb.Email,
		IsChirpyRed: userDb.IsChirpyRed,
	}

	respondWithJSON(w, http.StatusCreated, user)
}

func (cfg *apiConfig) updateUserHandler(w http.ResponseWriter, req *http.Request) {
	reqBody := Auth{}

	_ = unmarshalType(req, &reqBody)

	stringToken, err := auth.GetBearerToken(req.Header)
	if err != nil {
		log.Printf("Unable to get the token from request header: %s %s [%s]", req.Method, req.URL.Path, err)
		respondWithError(w, http.StatusUnauthorized, "")
		return
	}

	UserID, err := auth.ValidateJWT(stringToken, cfg.secret)
	if err != nil {
		log.Printf("Unable to validate the token: %s %s [%s]", req.Method, req.URL.Path, err)
		respondWithError(w, http.StatusUnauthorized, "")
		return
	}

	if len(strings.Split(reqBody.Email, "@")) < 2 {
		log.Printf("Invalid e-mail: %s", reqBody.Email)
		return
	}

	reqBody.Password, err = auth.HashPassword(reqBody.Password)
	if err != nil {
		log.Printf("Unable to hash the password: %s", err)
		return
	}

	updateUser := database.UpdateUserParams{
		Email:          reqBody.Email,
		HashedPassword: reqBody.Password,
		ID:             UserID,
	}

	userDb, err := cfg.dbQueries.UpdateUser(req.Context(), updateUser)
	if err != nil {
		log.Printf("Unable to update user e-mail and password: %s %s [%s]", req.Method, req.URL.Path, err)
		return
	}

	user := User{
		ID:          userDb.ID,
		CreatedAt:   userDb.CreatedAt,
		UpdatedAt:   userDb.UpdatedAt,
		Email:       userDb.Email,
		IsChirpyRed: userDb.IsChirpyRed,
	}

	respondWithJSON(w, http.StatusOK, user)
}

func (cfg *apiConfig) userUpgradeHandler(w http.ResponseWriter, req *http.Request) {
	reqBody := Event{}

	_ = unmarshalType(req, &reqBody)
	if reqBody.Event != "user.upgraded" {
		log.Printf("Unknown event: %s %s [%s]", req.Method, req.URL.Path, reqBody.Event)
		respondWithJSON(w, http.StatusNoContent, "")
	}

	/*log.Print("####################")
	log.Print(reqBody.Data.User_id)
	log.Print("####################")*/
	userID, err := uuid.Parse(reqBody.Data.User_id)
	if err != nil {
		log.Printf("Unable to parse userID: %s", reqBody.Data.User_id)
		return
	}

	userDb, err := cfg.dbQueries.UpgradeUser(req.Context(), userID)
	log.Print("####################")
	log.Printf("user - %s, Red status - %v : %v", userDb.Email, userDb.IsChirpyRed, err)
	log.Print("####################")
	if err == sql.ErrNoRows {
		log.Printf("User not found")
		respondWithError(w, http.StatusNotFound, "")
		return
	} else if err != nil {
		log.Printf("Unable to upgrade user: %s", userDb.ID)
		respondWithError(w, http.StatusInternalServerError, "")
		return
	}

	respondWithJSON(w, http.StatusNoContent, "")
}

func (cfg *apiConfig) loginHandler(w http.ResponseWriter, req *http.Request) {
	reqBody := Auth{}

	_ = unmarshalType(req, &reqBody)

	userDb, err := cfg.dbQueries.GetUserByEmail(req.Context(), reqBody.Email)
	if err != nil {
		log.Printf("Unable to retrieve user with the e-mail: %s %s [%s]", req.Method, req.URL.Path, reqBody.Email)
		return
	}

	err = auth.CheckPasswordHash(userDb.HashedPassword, reqBody.Password)
	if err != nil {
		log.Printf("Incorrect email or password: %s %s [%s]", req.Method, req.URL.Path, err)
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password")
		return
	}

	token, err := auth.MakeJWT(userDb.ID, cfg.secret, cfg.authExpiry)
	if err != nil {
		log.Printf("Unable to create token for user: %s", userDb.ID)
		return
	}

	refreshToken, _ := auth.MakeRefreshToken()
	refreshTokenDb, err := cfg.dbQueries.StoreRefreshToken(req.Context(),
		database.StoreRefreshTokenParams{Token: refreshToken, UserID: userDb.ID},
	)
	if err != nil {
		log.Printf("Unable to store new refresh token %s %s [%s]", req.Method, req.URL.Path, err)
		return
	}

	user := User{
		ID:          userDb.ID,
		CreatedAt:   userDb.CreatedAt,
		UpdatedAt:   userDb.UpdatedAt,
		Email:       userDb.Email,
		Token:       token,
		Refresh:     refreshTokenDb.Token,
		IsChirpyRed: userDb.IsChirpyRed,
	}

	respondWithJSON(w, http.StatusOK, user)
}

func (cfg *apiConfig) createChirpHandler(w http.ResponseWriter, req *http.Request) {
	reqBody := Chirp{}

	_ = unmarshalType(req, &reqBody)

	stringToken, err := auth.GetBearerToken(req.Header)
	if err != nil {
		log.Printf("Unable to get the token from request header: %s %s [%s]", req.Method, req.URL.Path, err)
		return
	}

	reqBody.UserID, err = auth.ValidateJWT(stringToken, cfg.secret)
	if err != nil {
		log.Printf("Unable to validate the token: %s %s [%s]", req.Method, req.URL.Path, err)
		respondWithError(w, http.StatusUnauthorized, "")
		return
	}

	if len(reqBody.Body) <= 140 {
		reqBody.Body = wordFilter(reqBody.Body)
		chirpDb, err := cfg.dbQueries.CreateChirp(req.Context(),
			database.CreateChirpParams{Body: reqBody.Body, UserID: reqBody.UserID})
		if err != nil {
			log.Printf("Unable to create chirp: %s %s [%s]", req.Method, req.URL.Path, err)
			respondWithError(w, http.StatusBadRequest, err.Error())
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
	if err == sql.ErrNoRows {
		log.Printf("No chirps not found")
		respondWithError(w, http.StatusNotFound, "")
		return
	} else if err != nil {
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
	if err == sql.ErrNoRows {
		log.Printf("Chirp not found")
		respondWithError(w, http.StatusNotFound, "")
		return
	} else if err != nil {
		log.Printf("Unable to retrieve chirp")
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

func (cfg *apiConfig) deleteChirpHandler(w http.ResponseWriter, req *http.Request) {
	reqBody := Auth{}

	_ = unmarshalType(req, &reqBody)

	stringToken, err := auth.GetBearerToken(req.Header)
	if err != nil {
		log.Printf("Unable to get the token from request header: %s %s [%s]", req.Method, req.URL.Path, err)
		respondWithError(w, http.StatusUnauthorized, "")
		return
	}

	UserID, err := auth.ValidateJWT(stringToken, cfg.secret)
	if err != nil {
		log.Printf("Unable to validate the token: %s %s [%s]", req.Method, req.URL.Path, err)
		respondWithError(w, http.StatusUnauthorized, "")
		return
	}

	chirpID, err := uuid.Parse(req.PathValue("chirpID"))
	if err != nil {
		log.Printf("Unable to parse chirpID: %s", req.PathValue("chirpID"))
		return
	}

	chirpDb, err := cfg.dbQueries.GetChirpById(req.Context(), chirpID)
	if err == sql.ErrNoRows {
		log.Printf("Chirp not found")
		respondWithError(w, http.StatusNotFound, "")
		return
	} else if err != nil {
		log.Printf("Unable to retrieve chirp: %s", chirpDb.ID)
		respondWithError(w, http.StatusInternalServerError, "")
		return
	}

	if chirpDb.UserID != UserID {
		log.Printf("Unable to validate user")
		respondWithError(w, http.StatusForbidden, "")
		return
	}

	err = cfg.dbQueries.DeleteChirpById(req.Context(), chirpID)
	if err != nil {
		log.Printf("Unable to delete chirp")
		return
	}

	respondWithJSON(w, http.StatusNoContent, "")
}

func (cfg *apiConfig) refreshHandler(w http.ResponseWriter, req *http.Request) {
	stringRefreshToken, err := auth.GetBearerToken((req.Header))
	if err != nil {
		log.Printf("Unable to get the token from request header: %s %s [%s]", req.Method, req.URL.Path, err)
		return
	}

	userID, err := cfg.dbQueries.GetUserFromRefreshToken(req.Context(), stringRefreshToken)
	if err != nil {
		log.Printf("Unable to get user by refresh token: %s %s [%s]", req.Method, req.URL.Path, err)
		respondWithError(w, http.StatusUnauthorized, "Refresh token expired or does not exist")
		return
	}

	newAuthToken, err := auth.MakeJWT(userID, cfg.secret, cfg.authExpiry)
	if err != nil {
		log.Printf("Unable to create token for user: %s", userID)
		return
	}
	respondWithJSON(w, http.StatusOK, struct {
		Token string `json:"token"`
	}{Token: newAuthToken})

}

func (cfg *apiConfig) revokeHandler(w http.ResponseWriter, req *http.Request) {
	stringRefreshToken, err := auth.GetBearerToken((req.Header))
	if err != nil {
		log.Printf("Unable to get the token from request header: %s %s [%s]", req.Method, req.URL.Path, err)
		return
	}

	err = cfg.dbQueries.RevokeRefershToken(req.Context(), stringRefreshToken)
	if err != nil {
		log.Printf("Unable to revoke the token from request header: %s %s [%s]", req.Method, req.URL.Path, err)
		respondWithError(w, http.StatusInternalServerError, "Token revokation unsuccessful")
		return
	}

	respondWithJSON(w, http.StatusNoContent, "")
}
