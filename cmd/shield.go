package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var shieldcmd = &cobra.Command{
	Use:   "shield",
	Short: "s",
	Long:  `shield is the cache layer api`,
	RunE: func(cmd *cobra.Command, args []string) error {
		var logger *zap.Logger

		debug, err := cmd.Flags().GetBool("debug-mode")
		if err != nil {
			return err
		}

		logger, err = zap.NewDevelopment()
		if err != nil {
			return err
		}

		logger.Info("mode", zap.Bool("debug", debug))

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

		return s.ListenAndServe()
	},
}

func init() {
	cobra.OnInitialize(initConfig)
	shieldcmd.Flags().Bool("debug-mode", false, "debug mode")
	rootCmd.AddCommand(shieldcmd)
}
