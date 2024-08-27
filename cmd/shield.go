package cmd

import (
	"net/http"
	"net/http/httputil"
	"nginy/cmd/shield"
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

		// TODO: fix this!
		logger, err = zap.NewDevelopment()
		if err != nil {
			return err
		}

		logger.Info("mode", zap.Bool("debug", debug))

		if err != nil {
			panic(err)
		}

		director := shield.NewDirector("http://localhost:1234")
		response := shield.NewResponse()

		mux := http.NewServeMux()
		proxy := &httputil.ReverseProxy{
			Director:       director.Request(),
			ModifyResponse: response.Modify(),
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
