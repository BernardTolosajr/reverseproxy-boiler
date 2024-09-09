package shield

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/open-feature/go-sdk/openfeature"
	"go.opentelemetry.io/otel/attribute"
	api "go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/sdk/metric"
)

type CacheMiddleware struct {
	api.Float64Counter
	api.Float64Histogram
	feature *openfeature.Client
}

func NewCacheMiddleware(exporter metric.Reader, feature *openfeature.Client) *CacheMiddleware {
	provider := metric.NewMeterProvider(metric.WithReader(exporter))
	meter := provider.Meter("shield")

	counter, err := meter.Float64Counter("counter", api.WithDescription("counter succes rate"))
	if err != nil {
		panic(err)
	}

	histogram, err := meter.Float64Histogram(
		"duration",
		api.WithDescription("a shield buckets"),
		api.WithExplicitBucketBoundaries(64, 128, 256, 512, 1024, 2048, 4096),
	)

	if err != nil {
		panic(err)
	}

	return &CacheMiddleware{
		counter,
		histogram,
		feature,
	}
}

func (c *CacheMiddleware) Next(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		w.Header()

		res, err := c.feature.BooleanValueDetails(context.Background(),
			"whitelist",
			false,
			openfeature.NewEvaluationContext(
				"<replace this msisdn here or any key>",
				map[string]interface{}{
					"key": "name:", // this is the key in the bucket
				},
			),
		)

		if err != nil {
			// TODO: log error using zap
			fmt.Errorf("%v", err)
		}

		if res.Value {
			// meaning we found payload in the whitelist, skip your logic!
		}

		opt := api.WithAttributes(
			attribute.Key("reader").String("success"),
		)

		// TODO: count success rate
		c.Add(context.Background(), 1, opt)

		// TODO: record duration
		c.Record(context.Background(), 136, opt)

		//TODO: add read cache logic
		_, err = json.Marshal(map[string]string{"hello": "world"})
		if err != nil {
			panic(err)
		}

		//w.Write([]byte(jsonString))

		next.ServeHTTP(w, r)
		fmt.Println(r.Method, r.URL.Path, time.Since(start))
	})
}
