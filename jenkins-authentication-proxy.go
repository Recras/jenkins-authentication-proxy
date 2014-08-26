package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"

	flag "github.com/docker/docker/pkg/mflag"
)

var jenkins_address string
var listen_address string

func main() {
	flag.StringVar(&jenkins_address, []string{"-jenkins", "j"}, "http://localhost:80", "The address Jenkins is running on")
	flag.StringVar(&listen_address, []string{"-listen", "l"}, "[::]:8080", "The address to listen on")
	flag.Parse()

	remote, err := url.Parse(jenkins_address)
	if err != nil {
		log.Panic(err)
	}

	proxy := httputil.NewSingleHostReverseProxy(remote)
	http.HandleFunc("/", handler(proxy))

	err = http.ListenAndServe(listen_address, nil)
	if err != nil {
		log.Panic(err)
	}
}

func handler(fw *httputil.ReverseProxy) func(http.ResponseWriter, *http.Request) {
	return func(wr http.ResponseWriter, req *http.Request) {
		var r *http.Request
		var err error
		var resp *http.Response
		client := &http.Client{}

		url := "https://recras.plan.io/users/current.json"
		r, err = http.NewRequest("GET", url, nil)
		if err != nil {
			wr.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(wr, "error", err)
			return
		}
		r.Header["Authorization"] = req.Header["Authorization"]

		resp, err = client.Do(r)
		if err != nil {
			wr.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(wr, "error", err)
			return
		}

		if resp.StatusCode == 200 {
			resp.Body.Close()
			fw.ServeHTTP(wr, req)
			return
		} else {
			wr.Header()["Www-Authenticate"] = resp.Header["Www-Authenticate"]
			wr.WriteHeader(http.StatusUnauthorized)
		}
	}
}
