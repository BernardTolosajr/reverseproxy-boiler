package shield

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type Director struct {
	origin string
}

func NewDirector(origin string) *Director {
	return &Director{
		origin,
	}
}

func (d *Director) Request() func(req *http.Request) {
	return func(req *http.Request) {
		remote, err := url.Parse(d.origin)
		if err != nil {
			panic(err)
		}
		req.Header.Add("X-Forwarded-Host", remote.Host)
		req.Header.Add("X-Origin-Host", remote.Host)
		req.URL.Scheme = remote.Scheme
		req.URL.Host = remote.Host

		path := fmt.Sprintf("%s/%s", remote.Path, strings.TrimLeft(req.URL.Path, "/"))

		req.URL.Path = path
	}
}
