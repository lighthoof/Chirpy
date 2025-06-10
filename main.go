package main

import (
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
