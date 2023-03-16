package api

import (
	"log"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	startTime time.Time
)

func InitAndListen() error {
	startTime = time.Now()

	log.Printf("Setting up http api...")
	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/health", getHealth)
	http.HandleFunc("/state", getState)
	http.HandleFunc("/cars", getCars)

	log.Printf("Listening for requests on %s.", "0.0.0.0:2112")
	return http.ListenAndServe(":2112", nil)
}
