package shield

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
)

type Response struct {
}

func NewResponse() *Response {
	return &Response{}
}

func (r *Response) Modify() func(*http.Response) error {
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
