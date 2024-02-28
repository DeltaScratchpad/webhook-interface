package server

import (
	"context"
	"errors"
	"fmt"
	"github.com/DeltaScratchpad/webhook-interface/helpers"
	"github.com/DeltaScratchpad/webhook-interface/processing"
	"github.com/DeltaScratchpad/webhook-interface/webhook-tracker"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

func CreateServer(addr *string, port string, done <-chan os.Signal, state webhook_tracker.WebhookState) {
	var handler = WebhookQueryHandler{
		webhookState: state,
		waitGroup:    new(sync.WaitGroup),
	}

	//Create request multiplexer
	mux := http.NewServeMux()

	//Add handler to multiplexer
	mux.Handle("/query", &handler)
	mux.Handle("/query/", &handler)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	if addr != nil {
		port = fmt.Sprintf("%s:%s", *addr, port)
	} else {
		port = fmt.Sprintf(":%s", port)
	}

	srv := &http.Server{
		Addr:    port,
		Handler: mux,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("listen: %s\n", err)
		}
	}()
	log.Println("Starting server on port " + port)

	// Wait for interrupt signal to gracefully shutdown the server with
	<-done
	log.Println("Server Stopping")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		// extra handling here
		cancel()
		handler.waitGroup.Wait()
		log.Println("Server Stopped")
	}()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server Shutdown Failed:%+v", err)
	}
}

type WebhookQueryHandler struct {
	webhookState webhook_tracker.WebhookState
	waitGroup    *sync.WaitGroup
}

func (q *WebhookQueryHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		q.HandleQuery(w, r)
		return
	default:
		helpers.InvalidHttpMethodHandler(w, r)
		return
	}
}

func (q *WebhookQueryHandler) HandleQuery(w http.ResponseWriter, r *http.Request) {
	q.waitGroup.Add(1) // Increment the wait group to ensure the event is forwarded before quitting.
	//Parse query
	query, err := helpers.ParseProcessingEvent(w, r)
	if err != nil {
		log.Printf("Error parsing query: %s \n", err)
		helpers.InternalServerErrorHandler(w, r)
		return
	}

	//log.Println("Handling query")
	defer func() { // Ensure the event will be forwarded regardless of errors.
		go func() {
			helpers.ForwardEvent(&query)
			// Don't let the server exit until the event has been forwarded.
			q.waitGroup.Done()
		}()
	}()
	processing.ProcessProcessingEvent(&query, q.webhookState)
}
