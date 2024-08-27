package shield

import "net/http"

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
		req.Header.Add("X-Forwarded-Host", d.origin)
		req.Header.Add("X-Origin-Host", d.origin)
		req.URL.Scheme = "http"
		req.URL.Host = d.origin
	}
}
