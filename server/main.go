package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/gorilla/mux"

	"EHDW/Cubic-spline-interpolation/server/api"
)

var (
	defaultServerAddress   = ":8000"
	defaultShutdownTimeout = time.Second * 20
)

func main() {
	// init values
	// create router
	router := mux.NewRouter()
	api.RegisterRoutes(router)

	path, _ := os.Getwd()
	path = filepath.Join(path, "public")

	api.AttachSPA(router, path, path)
	// adapt router with middleware
	router.Use(mux.CORSMethodMiddleware(router))
	// create http server
	server := &http.Server{
		Addr:        defaultServerAddress,
		ReadTimeout: time.Minute * 2,
		Handler:     router,
	}

	// launch server
	go func() {
		if err := server.ListenAndServe(); err != nil {
			log.Println(err)
		}
	}()
	log.Println("The cubic spline interpolation server is lanched successfuly! Visit localhost:8000 to use the app")

	// prepare graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-c

	// hanlde shutdown
	ctx, cancel := context.WithTimeout(context.Background(), defaultShutdownTimeout)
	defer cancel()

	server.Shutdown(ctx)

	log.Println("Shutting down...")
	os.Exit(0)
}
