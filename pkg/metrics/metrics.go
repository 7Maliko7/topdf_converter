package metrics

import "github.com/prometheus/client_golang/prometheus"

var (
	UserCount = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "users_count",
			Help: "The total number of users",
		},
	)
)

var (
	PhotoCount = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "photo_count",
			Help: "The total number of uploaded photos",
		},
	)
)

var (
	PdfCount = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "pdf_count",
			Help: "The total number of pdf",
		},
	)
)

var (
	UserInteractionDuration = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "interaction_duration",
			Help:    "Histogram of interacrion durations with the bot in seconds",
			Buckets: prometheus.DefBuckets,
		},
	)
)
