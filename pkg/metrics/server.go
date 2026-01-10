package metrics

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func StartMetricsServer(ctx context.Context) error {
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

		shutdownCtx, cancel := context.WithTimeout(
			context.Background(),
			5*time.Second,
		)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("prometheus shutdown error: %v", err)
		}
	}()

	log.Println("Starting Preometheus metrics server on port 9090")
	err := server.ListenAndServe()
	if err == http.ErrServerClosed {
		return nil
	}
	return err
}

func registerMetcrics() {
	prometheus.MustRegister(UserCount)
	prometheus.MustRegister(PhotoCount)
	prometheus.MustRegister(PdfCount)
	prometheus.MustRegister(UserInteractionDuration)

}
