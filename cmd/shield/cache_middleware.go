package shield

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type CacheMiddleware struct {
}

func NewCacheMiddleware() *CacheMiddleware {
	return &CacheMiddleware{}
}

func (c *CacheMiddleware) Next(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		w.Header()

		// read cache logic

		_, err := json.Marshal(map[string]string{"hello": "world"})
		if err != nil {
			panic(err)
		}

		//w.Write([]byte(jsonString))

		next.ServeHTTP(w, r)
		fmt.Println(r.Method, r.URL.Path, time.Since(start))
	})
}
