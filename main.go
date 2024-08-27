package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"strconv"
	"time"
)

func ModifiedResponse() func(*http.Response) error {
	return func(r *http.Response) error {
		var body map[string]interface{}

		err := json.NewDecoder(r.Body).Decode(&body)
		if err != nil {
			return err
		}

		fmt.Printf("%v", body)

		// write cache response here

		jsonString, err := json.Marshal(body)
		if err != nil {
			return err
		}

		r.Body = ioutil.NopCloser(bytes.NewBuffer(jsonString))
		r.ContentLength = int64(len(jsonString))
		r.Header.Set("Content-Length", strconv.Itoa(len(jsonString)))
		r.Header.Set("Content-Type", "application/json")

		return nil
	}
}

func ErrorHandler() func(http.ResponseWriter, *http.Request, error) {
	return func(w http.ResponseWriter, req *http.Request, err error) {
		fmt.Printf("Got error while modifying response: %v \n", err)
		return
	}
}

func CacheMiddleware(next http.Handler) http.Handler {
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

func main() {
	// mock server
	be := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("response from the server")
		w.WriteHeader(http.StatusCreated)
		w.Header().Set("Content-Type", "application/json")
		resp := make(map[string]string)
		resp["message"] = "hello there!"
		jsonResp, err := json.Marshal(resp)
		if err != nil {
			log.Fatalf("Error happened in JSON marshal. Err: %s", err)
		}
		w.Write(jsonResp)
	}))
	defer be.Close()

	origin, err := url.Parse(be.URL)
	if err != nil {
		panic(err)
	}

	director := func(req *http.Request) {
		req.Header.Add("X-Forwarded-Host", req.Host)
		req.Header.Add("X-Origin-Host", origin.Host)
		req.URL.Scheme = "http"
		req.URL.Host = origin.Host
	}

	mux := http.NewServeMux()

	proxy := &httputil.ReverseProxy{
		Director:       director,
		ModifyResponse: ModifiedResponse(),
		ErrorHandler:   ErrorHandler(),
	}

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		proxy.ServeHTTP(w, r)
	})

	s := &http.Server{
		Addr:           ":9000",
		Handler:        CacheMiddleware(mux),
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	log.Fatal(s.ListenAndServe())
}
