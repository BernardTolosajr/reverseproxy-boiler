package shield

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go.opentelemetry.io/otel/attribute"
	api "go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/sdk/metric"
)

type CacheMiddleware struct {
	counter   api.Float64Counter
	histogram api.Float64Histogram
}

func NewCacheMiddleware(exporter metric.Reader) *CacheMiddleware {
	provider := metric.NewMeterProvider(metric.WithReader(exporter))
	meter := provider.Meter("shield")

	counter, err := meter.Float64Counter("counter", api.WithDescription("counter succes rate"))
	if err != nil {
		panic(err)
	}

	histogram, err := meter.Float64Histogram(
		"duration",
		api.WithDescription("a shieldbuckets"),
		api.WithExplicitBucketBoundaries(64, 128, 256, 512, 1024, 2048, 4096),
	)

	if err != nil {
		panic(err)
	}

	return &CacheMiddleware{
		counter,
		histogram,
	}
}

func (c *CacheMiddleware) Next(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		w.Header()

		opt := api.WithAttributes(
			attribute.Key("reader").String("success"),
		)

		// TODO: count success rate
		c.counter.Add(context.Background(), 1, opt)

		// TODO: record duration
		c.histogram.Record(context.Background(), 136, opt)

		//TODO: add read cache logic
		_, err := json.Marshal(map[string]string{"hello": "world"})
		if err != nil {
			panic(err)
		}

		//w.Write([]byte(jsonString))

		next.ServeHTTP(w, r)
		fmt.Println(r.Method, r.URL.Path, time.Since(start))
	})
}
