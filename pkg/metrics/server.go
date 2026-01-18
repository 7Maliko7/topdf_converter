package metrics

import (
	"context"
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func StartMetricsServer(ctx context.Context, errChan chan error) {
	registerMetcrics()

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	server := &http.Server{
		Addr:    ":9090",
		Handler: mux,
	}

	go func() {
		<-ctx.Done()
		log.Println("stopping prometheus server")

		if err := server.Shutdown(ctx); err != nil {
			log.Printf("prometheus shutdown error: %v", err)
			errChan <- err
		}
	}()

	log.Println("Starting Preometheus metrics server on port 9090")

	go func() {
		err := server.ListenAndServe()
		if err != nil {
			errChan <- err
		}
	}()
}

func registerMetcrics() {
	prometheus.MustRegister(UserCount)
	prometheus.MustRegister(PhotoCount)
	prometheus.MustRegister(PdfCount)
	prometheus.MustRegister(UserInteractionDuration)

}
