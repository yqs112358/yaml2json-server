package main

import (
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/bronze1man/yaml2json/y2jLib"
)

var (
	listenAddr string
	listenPort int
	authKey    string
	urlSubPath string
	Version    string
)

func parseConfigs() {
	var err error = nil

	// parse env vars
	if addr := os.Getenv("Y2JS_LISTEN_ADDR"); addr != "" {
		listenAddr = addr
	} else {
		listenAddr = "0.0.0.0"
	}

	if port := os.Getenv("Y2JS_LISTEN_PORT"); port != "" {
		listenPort, err = strconv.Atoi(port)
		if err != nil {
			log.Fatalf("Bad port number: %s", port)
		}
	} else {
		listenPort = 8080
	}

	authKey = os.Getenv("Y2JS_AUTH_KEY")

	if subPath := os.Getenv("Y2JS_URL_SUB_PATH"); subPath != "" {
		urlSubPath = subPath
	} else {
		urlSubPath = "/"
	}

	// parse cmd-line args
	showVersion := false
	flag.StringVar(&listenAddr, "listen", listenAddr, "HTTP listen address")
	flag.IntVar(&listenPort, "port", listenPort, "HTTP listen port")
	flag.StringVar(&authKey, "key", authKey, "Pre-shared auth key")
	flag.StringVar(&urlSubPath, "sub-path", urlSubPath, "HTTP Serve sub-path")
	flag.BoolVar(&showVersion, "version", false, "Print the version and exit")
	flag.Parse()

	// check configs
	if showVersion {
		if Version == "" {
			Version = "development"
		}
		fmt.Println("yaml2json-server Version:", Version)
		os.Exit(0)
	}
	if listenPort < 0 || listenPort > 65535 {
		log.Fatalf("Bad port number: %d", listenPort)
	}
}

func checkAuth(r *http.Request, key string) bool {
	// check key in URL
	if r.URL.Query().Get("key") == key {
		return true
	}

	// check http basic auth
	auth := r.Header.Get("Authorization")
	if auth != "" && strings.HasPrefix(auth, "Basic ") {
		payload, _ := base64.StdEncoding.DecodeString(strings.TrimPrefix(auth, "Basic "))
		if string(payload) == ":"+key {
			return true
		}
	}
	return false
}

func httpReturnError(w http.ResponseWriter, statusCode int, reason string) {
	w.WriteHeader(statusCode)
	errorResponse := map[string]string{"error": reason}
	_ = json.NewEncoder(w).Encode(errorResponse)
}

func httpHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// authentication
	if authKey != "" {
		if !checkAuth(r, authKey) {
			httpReturnError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}
	}

	// read YAML from given URL
	urlParam := r.URL.Query().Get("url")
	if urlParam != "" {
		resp, err := http.Get(urlParam)
		if err != nil {
			httpReturnError(w, http.StatusBadRequest, "Failed to fetch YAML from given URL")
			return
		}
		defer func(Body io.ReadCloser) {
			_ = Body.Close()
		}(resp.Body)

		err = y2jLib.TranslateStream(resp.Body, w)
		if err != nil {
			httpReturnError(w, http.StatusInternalServerError, "Failed to convert YAML to JSON")
			return
		}
	} else {
		// or, parse YAML from request body
		if r.ContentLength == 0 {
			httpReturnError(w, http.StatusBadRequest, "No YAML to convert")
			return
		}
		err := y2jLib.TranslateStream(r.Body, w)
		if err != nil {
			httpReturnError(w, http.StatusInternalServerError, "Failed to convert YAML to JSON")
			return
		}
	}
}

func main() {
	parseConfigs()

	http.HandleFunc(urlSubPath, httpHandler)

	addr := fmt.Sprintf("%s:%d", listenAddr, listenPort)
	log.Printf("yaml2json-server is listening on %s%s", addr, urlSubPath)
	log.Fatal(http.ListenAndServe(addr, nil))
}
