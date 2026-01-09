package metrics

import (
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func StartMetricsServer() {
	prometheus.MustRegister(UserCount)
	prometheus.MustRegister(PhotoCount)
	prometheus.MustRegister(PdfCount)
	prometheus.MustRegister(UserInteractionDuration)

	http.Handle("/metrics", promhttp.Handler())
	go func() {
		log.Println("Starting Preometheus metrics server on port 9090")
		if err := http.ListenAndServe(":9090", nil); err != nil {
			log.Fatalf("Failed to start metrics server: %v", err)
		}
	}()
}
