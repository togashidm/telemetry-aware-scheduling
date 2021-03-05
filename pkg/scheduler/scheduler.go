//Package scheduler extender logic contains code to respond call from the http endpoint.
package scheduler

import (
	"crypto/tls"
	"log"
	"net/http"
	"os"
	"time"
)

//postOnly check if the method type is POST
func postOnly(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			log.Print("method Type not POST")
			return
		}
		next.ServeHTTP(w, r)
	}
}

//contentLength check the if the request size is adequate
func contentLength(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.ContentLength > 1*1000*1000*1000 {
			w.WriteHeader(http.StatusInternalServerError)
			log.Print("request size too large")
			return
		}
		next.ServeHTTP(w, r)
	}
}

//requestContentType verify the content type of the request
func requestContentType(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requestContentType := r.Header.Get("Content-Type")
		if requestContentType != "application/json" {
			w.WriteHeader(http.StatusNotFound)
			log.Print("request size too large")
			return
		}
		next.ServeHTTP(w, r)
	}
}

//handlerWithMiddleware is handler wrapped with middleware to serve the prechecks at endpoint
func handlerWithMiddleware(handle http.HandlerFunc) http.HandlerFunc {
	return requestContentType(
		contentLength(
			postOnly(handle)))
}

//error handler deals with requests sent to an invalid endpoint and returns a 404.
func errorHandler(w http.ResponseWriter, r *http.Request) {
	log.Print("unknown path")
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotFound)
}

//Check symlinks checks if a file is a simlink and returns an error if it is.
func checkSymLinks(filename string) error {
	info, err := os.Lstat(filename)
	if err != nil {
		return err
	}
	if info.Mode() == os.ModeSymlink {
		return err
	}
	return nil
}

// StartServer starts the HTTP server needed for scheduler.
// It registers the handlers and checks for existing telemetry policies.
func (m Server) StartServer(port string, certFile string, keyFile string, unsafe bool) {
	http.HandleFunc("/", handlerWithMiddleware(errorHandler))
	http.HandleFunc("/scheduler/prioritize", handlerWithMiddleware(m.Prioritize))
	http.HandleFunc("/scheduler/filter", handlerWithMiddleware(m.Filter))
	var err error
	if unsafe {
		log.Printf("Extender Listening on HTTP  %v", port)
		err = http.ListenAndServe(":"+port, nil)
	} else {
		err := checkSymLinks(certFile)
		if err != nil {
			panic(err)
		}
		err = checkSymLinks(keyFile)
		if err != nil {
			panic(err)
		}
		log.Printf("Extender Now Listening on HTTPS  %v", port)
		srv := configureSecureServer(port)
		log.Fatal(srv.ListenAndServeTLS(certFile, keyFile))
	}
	log.Printf("Scheduler extender failed %v ", err)
}

//Configuration values including algorithms etc for the TAS scheduling endpoint.
func configureSecureServer(port string) *http.Server {
	cfg := &tls.Config{
		MinVersion:               tls.VersionTLS12,
		CurvePreferences:         []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256},
		PreferServerCipherSuites: true,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
		},
	}
	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           nil,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      10 * time.Second,
		MaxHeaderBytes:    1000,
		TLSConfig:         cfg,
		TLSNextProto:      make(map[string]func(*http.Server, *tls.Conn, http.Handler)),
	}
	return srv
}
