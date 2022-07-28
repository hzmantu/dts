package report

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
)

func init() {
	// Expose the registered metrics via HTTP.
	http.Handle("/metrics", promhttp.HandlerFor(
		prometheus.DefaultGatherer,
		promhttp.HandlerOpts{
			// Opt into OpenMetrics to support exemplars.
			EnableOpenMetrics: true,
		},
	))
}

var (
	ReaderCount = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mysql_reader",
			Help: "mysql_reader_count",
		}, []string{"database", "table", "source"},
	)
	ReaderLength = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mysql_reader_length",
			Help: "mysql_reader_length_count",
		}, []string{"database", "table", "source"},
	)
	DumpCount = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mysql_dump",
			Help: "mysql_dump_count",
		}, []string{"source", "database", "table"},
	)
	CompareCount = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mysql_compare",
			Help: "mysql_compare_count",
		}, []string{"database", "table", "operate"},
	)
)
