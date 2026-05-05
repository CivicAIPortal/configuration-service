package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

var loadConfigFn = getConfig
var getConfigMapFn = k8sGetConfigmap

func RequestLogger(targetMux http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		targetMux.ServeHTTP(w, r)

		Logger.Infow("",
			zap.String("method", string(r.Method)),
			zap.String("uri", string(r.RequestURI)),
			zap.Duration("duration", time.Since(start)),
		)
	})
}

func startServer(port *int) {
	router := buildRouter()

	portString := ":" + strconv.Itoa(*port)
	log.Fatal(http.ListenAndServe(portString, router))
}

func buildRouter() *mux.Router {
	initMetrics()

	router := mux.NewRouter().StrictSlash(true)
	router.Use(metricsMiddleware)
	router.Use(RequestLogger)
	router.HandleFunc("/configuration", rolesGet).Methods("GET")
	router.HandleFunc("/health", healthGet).Methods("GET")
	router.Handle("/metrics", promhttp.Handler()).Methods("GET")
	router.HandleFunc("/isAlive", isAliveGet).Methods("GET")
	return router
}

func rolesGet(w http.ResponseWriter, r *http.Request) {
	// Get config
	config, _ := loadConfigFn()

	w.Header().Set("Content-Type", "application/json")

	// Get configmap's data
	data, err := getConfigMapFn(config.configmapName, config.configmapNamespace)
	if err != nil {
		Logger.Error(err)
		w.WriteHeader(500)
		return
	}

	json.NewEncoder(w).Encode(data)

	return
}

func isAliveGet(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)

	return
}
