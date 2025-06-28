package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/lighthoof/Chirpy/internal/database"
)

func main() {
	godotenv.Load()
	const filePathRoot = "."
	//const metricsTemplate = "./metrics_tmplt.html"
	const port = "8080"

	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Unable to open DB connection : %v", err)
	}

	cfg := apiConfig{
		fileserverHits: atomic.Int32{},
		dbQueries:      database.New(db),
		platform:       os.Getenv("PLATFORM"),
		secret:         os.Getenv("TOKEN_SECRET"),
		authExpiry:     time.Hour,
	}

	serveMux := http.NewServeMux()
	fileServerHandler := http.FileServer(http.Dir(filePathRoot))
	noPrefixFileHandler := http.StripPrefix("/app/", fileServerHandler)
	serveMux.Handle("/app/", middlewareLog(cfg.middlewareMetricsInc(noPrefixFileHandler)))
	//serveMux.HandleFunc("GET /api/metrics", conf.counterHandler)
	serveMux.HandleFunc("GET /admin/metrics", cfg.counterHandler)
	serveMux.HandleFunc("POST /admin/reset", cfg.resetHandler)
	//serveMux.HandleFunc("POST /api/validate_chirp", validationHandler)
	serveMux.HandleFunc("GET /api/healthz", readinessHandler)
	serveMux.HandleFunc("GET /api/chirps", cfg.getChirpsHandler)
	serveMux.HandleFunc("GET /api/chirps/{chirpID}", cfg.getChirpByIdHandler)
	serveMux.HandleFunc("POST /api/chirps", cfg.createChirpHandler)
	serveMux.HandleFunc("POST /api/users", cfg.createUserHandler)
	serveMux.HandleFunc("POST /api/login", cfg.loginHandler)
	serveMux.HandleFunc("POST /api/refresh", cfg.refreshHandler)
	serveMux.HandleFunc("POST /api/revoke", cfg.revokeHandler)
	serveMux.HandleFunc("PUT /api/users", cfg.updateUserHandler)
	serveMux.HandleFunc("DELETE /api/chirps/{chirpID}", cfg.deleteChirpByIdHandler)

	server := &http.Server{
		Handler: serveMux,
		Addr:    ":" + port,
	}

	log.Printf("Serving files from %s on port: %s\n", filePathRoot, port)
	err = server.ListenAndServe()
	if err != nil {
		log.Fatalf("Server error : %v", err)
	}
}

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
	Token     string    `json:"token"`
	Refresh   string    `json:"refresh_token"`
}
type Auth struct {
	Password string `json:"password"`
	Email    string `json:"email"`
}

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}
