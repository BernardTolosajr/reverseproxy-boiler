package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

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

var rootCmd = &cobra.Command{
	Use:   "shield",
	Short: "shield todo",
	Long: `A Fast and cache layer with
                love by btx and friends`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func initConfig() {
	viper.SetConfigFile(cfgFile)
	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.sample.yaml)")
}
