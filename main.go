package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"sync/atomic"
	"time"

	"github.com/su55y/geoguessr-cheat/internal/location"
)

type key int

const (
	requestIDKey key = 0
)

var (
	healthy   int32
	port      string
	geoMetaRX = regexp.MustCompile(
		"^\\S+req\\?u=(https:\\/\\/maps\\.googleapis\\.com\\/maps\\/api\\/js\\/GeoPhotoService\\.GetMetadata\\S+&callback=\\S+)$",
	)
)

func init() {
	flag.StringVar(&port, "p", "5000", "port")
	flag.Parse()
}

func main() {
	logger := log.New(os.Stdout, "http: ", log.LstdFlags)
	logger.Println("Server is starting...")

	router := http.NewServeMux()
	router.Handle("/req", geoReq(logger))

	nextRequestID := func() string {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      tracing(nextRequestID)(logging(logger)(router)),
		ErrorLog:     logger,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	done := make(chan bool)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)

	go func() {
		<-quit
		logger.Println("Server is shutting down...")
		atomic.StoreInt32(&healthy, 0)

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		server.SetKeepAlivesEnabled(false)
		if err := server.Shutdown(ctx); err != nil {
			logger.Fatalf("Could not gracefully shutdown the server: %v\n", err)
		}
		close(done)
	}()

	logger.Printf("Server is ready to handle requests at :%s", port)
	atomic.StoreInt32(&healthy, 1)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Fatalf("Could not listen on %s: %v\n", port, err)
	}

	<-done
	logger.Println("Server stopped")
}

func geoReq(logger *log.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", "application/json")
		if geoMetaRX.MatchString(r.URL.String()) {
			if u := geoMetaRX.FindStringSubmatch(r.URL.String()); u != nil && len(u) == 2 {
				if err := location.ProceedUrl(u[1], &w, logger); err != nil {
					logger.Printf("error while proceeding: %v", err.Error())
				}
			}
		}
	})
}

func logging(logger *log.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				requestID, ok := r.Context().Value(requestIDKey).(string)
				if !ok {
					requestID = "unknown"
				}
				logger.Printf(
					"request id:%s, url:%s\n",
					requestID,
					r.URL.String(),
				)
			}()
			next.ServeHTTP(w, r)
		})
	}
}

func tracing(nextRequestID func() string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := r.Header.Get("X-Request-Id")
			if requestID == "" {
				requestID = nextRequestID()
			}
			ctx := context.WithValue(r.Context(), requestIDKey, requestID)
			w.Header().Set("X-Request-Id", requestID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
