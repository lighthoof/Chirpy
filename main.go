package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
)

func main() {
	const filePathRoot = "."
	//const metricsTemplate = "./metrics_tmplt.html"
	const port = "8080"

	conf := apiConfig{
		fileserverHits: atomic.Int32{},
	}

	serveMux := http.NewServeMux()
	fileServerHandler := http.FileServer(http.Dir(filePathRoot))
	noPrefixFileHandler := http.StripPrefix("/app/", fileServerHandler)
	serveMux.Handle("/app/", middlewareLog(conf.middlewareMetricsInc(noPrefixFileHandler)))
	serveMux.HandleFunc("GET /api/healthz", readinessHandler)
	//serveMux.HandleFunc("GET /api/metrics", conf.counterHandler)
	serveMux.HandleFunc("GET /admin/metrics", conf.counterHandler)
	serveMux.HandleFunc("POST /admin/reset", conf.resetHandler)
	serveMux.HandleFunc("POST /api/validate_chirp", validationHandler)

	server := &http.Server{
		Handler: serveMux,
		Addr:    ":" + port,
	}

	log.Printf("Serving files from %s on port: %s\n", filePathRoot, port)
	err := server.ListenAndServe()
	if err != nil {
		log.Fatalf("Server error : %v", err)
	}
}

type apiConfig struct {
	fileserverHits atomic.Int32
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, req *http.Request) {
			cfg.fileserverHits.Add(1)
			next.ServeHTTP(w, req)
		},
	)
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
	data := []byte{}
	statusCode := http.StatusNotFound
	reqData := unmarshalJson(req)

	if len(reqData.Body) > 140 {
		errData := jsonError{JsonError: "Chirp is too long"}
		data, _ = json.Marshal(errData)
		statusCode = http.StatusBadRequest
	} else {
		errData := jsonError{JsonError: "Something went wrong"}
		data, _ = json.Marshal(errData)
		statusCode = http.StatusBadRequest
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write(data)
}

func middlewareLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}
