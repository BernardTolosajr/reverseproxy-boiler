package cmd

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"shield/cmd/shield"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"go.opentelemetry.io/otel/exporters/prometheus"
)

func ErrorHandler() func(http.ResponseWriter, *http.Request, error) {
	return func(w http.ResponseWriter, req *http.Request, err error) {
		fmt.Printf("Got error while modifying response: %v \n", err)
		return
	}
}

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

		// TODO: fix this!
		logger, err = zap.NewDevelopment()
		if err != nil {
			return err
		}

		logger.Info("mode", zap.Bool("debug", debug))

		if err != nil {
			panic(err)
		}

		origin, err := cmd.Flags().GetString("origin-host")
		if err != nil {
			return err
		}

		tm, err := cmd.Flags().GetBool("transparent-mode")
		if err != nil {
			return err
		}

		exporter, err := prometheus.New()
		if err != nil {
			return err
		}

		mux := http.NewServeMux()
		proxy := createProxy(origin, tm)

		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			proxy.ServeHTTP(w, r)
		})

		mux.Handle("/metrics", promhttp.Handler())

		var handler http.Handler

		mw := shield.NewCacheMiddleware(exporter)

		// transparent mode
		if !tm {
			handler = mw.Next(mux)
		} else {
			handler = mux
		}

		// TODO: support multiple middleware
		s := &http.Server{
			Addr:           ":9000",
			Handler:        handler,
			ReadTimeout:    10 * time.Second,
			WriteTimeout:   10 * time.Second,
			MaxHeaderBytes: 1 << 20,
		}

		return s.ListenAndServe()
	},
}

func createProxy(origin string, transparentMode bool) *httputil.ReverseProxy {
	director := shield.NewDirector(origin)

	if !transparentMode {
		response := shield.NewResponse()
		return &httputil.ReverseProxy{
			Director:       director.Request(),
			ModifyResponse: response.Modify(),
			ErrorHandler:   ErrorHandler(),
		}
	}

	return &httputil.ReverseProxy{
		Director:     director.Request(),
		ErrorHandler: ErrorHandler(),
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	shieldcmd.Flags().Bool("debug-mode", false, "debug mode")
	shieldcmd.Flags().String("origin-host", "http://localhost:3030", "origin host")
	shieldcmd.Flags().Bool("transparent-mode", false, "transparent mode")
	rootCmd.AddCommand(shieldcmd)
}
