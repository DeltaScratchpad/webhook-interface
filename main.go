package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/DeltaScratchpad/webhook-interface/helpers"
	webhook_tracker "github.com/DeltaScratchpad/webhook-interface/webhook-tracker"
	"log"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"sync"
	"syscall"
	"time"
)

func main() {
	var port = os.Getenv("PORT")
	if port == "" {
		port = "80"
	}

	//Create statistics actions
	var handler = WebhookQueryHandler{
		webhookState: webhook_tracker.NewLocalWebhookState(),
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
		return
	})

	srv := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	//Start server
	//log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), mux))
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

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

	defer func() { // Ensure the event will be forwarded regardless of errors.
		go func() {
			helpers.ForwardEvent(&query)
			// Don't let the server exit until the event has been forwarded.
			q.waitGroup.Done()
		}()
	}()

	//Parse args
	field, relation, threshold_int, threshold_str, webhook, err := parseArgs(&query.Commands.Commands[query.Commands.Step].Args)
	if err != nil {
		log.Printf("Error parsing args: %s \n", err)
		helpers.LogError(fmt.Sprintf("Failed to parse args: %s", err), &query)
		return
	}

	var result bool = false
	// Get field value
	value, err := helpers.GetIntValue(&query.Event, field)
	if err != nil {
		result = compareIntByRelation(value, threshold_int, relation)
	} else {
		value, err := helpers.GetStringValue(&query.Event, field)
		if err != nil {
			if relation == "=" {
				result = value == threshold_str
			}
		}
	}

	// Check if we should call it, but only if it hasn't already been called.
	// Under a race condition, multiple could be sent.
	//TODO: Would need distributed locking to resolve.
	if result && !q.webhookState.HasBeenCalled(webhook, query.Commands.QueryId) {
		err := helpers.SendGetWebhook(webhook)
		if err != nil {
			log.Printf("Error sending webhook: %s \n", err)
			q.webhookState.IncrementCallCount(webhook, query.Commands.QueryId)
			helpers.LogError(fmt.Sprintf("Failed to send webhook: %s", err), &query)
			return
		}
	}
	return
}

var args_parser = regexp.MustCompile(`(?P<field>\w+)(?P<relation>[><=]{1,2})(?P<threshold>\d+)\s(?P<webhook>.+)`)
var filed_index = args_parser.SubexpIndex(`field`)
var relation_index = args_parser.SubexpIndex(`relation`)
var threshold_index = args_parser.SubexpIndex(`threshold`)
var webhook_index = args_parser.SubexpIndex(`webhook`)

func parseArgs(args *string) (field string, relation string, threshold_int int, threshold_str string, webhook string, err error) {
	matches := args_parser.FindStringSubmatch(*args)

	field = matches[filed_index]
	relation = matches[relation_index]
	threshold_str = matches[threshold_index]
	threshold_int, err = strconv.Atoi(threshold_str)
	webhook = matches[webhook_index]
	if field == "" || relation == "" || threshold_str == "" || webhook == "" {
		err = fmt.Errorf("invalid args")
	}
	return
}

func compareIntByRelation(a int, b int, relation string) bool {
	switch relation {
	case ">":
		return a > b
	case ">=":
		return a >= b
	case "<":
		return a < b
	case "<=":
		return a <= b
	case "=":
		return a == b
	default:
		return false
	}
}
