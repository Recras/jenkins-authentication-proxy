package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
)

var openPrefixes = []string{
	"/git",
	"/buildByToken",
	"/cli",
	"/jnlpJars",
	"/subversion",
	"/whoAmI",
}

const planio_url = "https://recras.plan.io/users/current.json"

func main() {
	jenkins_address := os.Getenv("JENKINS_URL")
	listen_address := os.Getenv("LISTEN_ADDRESS")

	if listen_address == "" {
		listen_address = "[::]:8080"
	}
	if jenkins_address == "" {
		log.Fatalln("Use environment variables JENKINS_URL and LISTEN_ADDRESS (default \"[::]:8080\")")
	}

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

func isOpenPrefix(requestURI string) bool {
	for _, prefix := range openPrefixes {
		if strings.HasPrefix(requestURI, prefix) {
			return true
		}
	}
	return false
}

func authenticateWithBackend(req *http.Request) (bool, error) {
	var r *http.Request
	var err error
	var resp *http.Response

	r, err = http.NewRequest("GET", planio_url, nil)
	if err != nil {
		return false, err
	}
	r.Header["Authorization"] = req.Header["Authorization"]

	client := http.Client{}
	resp, err = client.Do(r)
	if err != nil {
		return false, err
	}

	resp.Body.Close()
	return resp.StatusCode == 200, nil
}

func handler(fw *httputil.ReverseProxy) func(http.ResponseWriter, *http.Request) {
	return func(wr http.ResponseWriter, req *http.Request) {
		if isOpenPrefix(req.RequestURI) {
			fw.ServeHTTP(wr, req)
			return
		}

		if authed, err := authenticateWithBackend(req); err != nil {
			wr.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(wr, "error", err)
			log.Print(err)
		} else if authed {
			fw.ServeHTTP(wr, req)
		} else {
			wr.Header()["Www-Authenticate"] = []string{"Basic realm=\"Jenkins\""}
			wr.WriteHeader(http.StatusUnauthorized)
		}
	}
}
