package main

import (
	"context"
	"encoding/json"
	"errors"
	go_system_api "github.com/DeltaScratchpad/go-system-api"
	"github.com/DeltaScratchpad/webhook-interface/server"
	webhook_tracker "github.com/DeltaScratchpad/webhook-interface/webhook-tracker"
	"log"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

const ListenServerPort = "9099"

var addr = "localhost"

func simpleWebhookListener(c chan<- bool, port string) {
	// This function creates a simple webserver which listens for an incoming web request.
	// When a request is received, the function sends a message to the channel.

	// Make a shutdown channel
	shutdown := make(chan bool)

	// Create a simple server
	mux := http.NewServeMux()
	mux.HandleFunc("/webhook", func(w http.ResponseWriter, r *http.Request) {
		c <- true
		shutdown <- true
	})

	srv := &http.Server{
		Addr:    addr + ":" + port,
		Handler: mux,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// Wait for the shutdown signal
	<-shutdown
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server Shutdown Failed:%+v", err)
	}
}

func TestHandlerServer(t *testing.T) {
	var server_port = "9099"
	var target_listener = "9100"
	var dump_listener = "9101"

	t.Log("Testing for a successful comparison.")

	// This function tests the WebhookQueryHandler, by creating an instance of it, then sending an example request
	// to the SimpleWebHook Listener function. The WebhookQueryHandler should receive the request and process it.

	// Create a channel to receive the message from the SimpleWebhookListener
	c := make(chan bool, 1)
	done := make(chan os.Signal, 1)
	dump := make(chan bool, 1)

	go server.CreateServer(&addr, server_port, done, webhook_tracker.NewLocalWebhookState())
	go simpleWebhookListener(c, target_listener)
	go simpleWebhookListener(dump, dump_listener)

	// Wait for 2 seconds to allow the servers to start
	time.Sleep(2 * time.Second)

	// Make test data
	raw := "Test"
	timeVal := time.Now()
	derived := make(map[string]interface{})
	derived["fieldname"] = 30
	var testEventData = go_system_api.EventData{
		Raw:       &raw,
		IndexTime: &timeVal,
		TimeStamp: &timeVal,
		EventType: &raw,
		Category:  &raw,
		Derived:   derived,
	}

	var testCommandList = go_system_api.CommandList{
		QueryId: "Test Command 1",
		Commands: []go_system_api.CommandStep{
			{CommandName: "Command 1", Args: "fieldname>=30 http://localhost:" + target_listener + "/webhook", Url: "http://localhost:" + server_port + "/query"},
			{CommandName: "Command 2", Args: "args 2", Url: "http://localhost:" + "9101" + "/webhook"},
		},
		Step:     0,
		ErrorUrl: "Error Url",
	}

	var testProcessingEvent = go_system_api.ProcessingEvent{
		Commands: testCommandList,
		Event:    testEventData,
	}

	log.Println("Sending test event")
	// Send the request
	err := sendEvent(&testProcessingEvent, "http://localhost:"+server_port+"/query")
	if err != nil {
		log.Fatalf("Error sending test event: %s", err)
	}

	log.Println("Waiting for webhook to be received")

	// Wait for the message from the SimpleWebhookListener
	select {
	case <-c:
		log.Println("Test Passed")
		return
	case <-time.After(5 * time.Second):
		log.Fatalf("Test Failed")
		t.FailNow()
	}
}

func TestShouldNotSendWebhook(t *testing.T) {
	var server_port = "9102"
	var target_listener = "9103"
	var dump_listener = "9104"

	t.Log("Testing for a negative comparison.")

	// This function tests the WebhookQueryHandler, by creating an instance of it, then sending an example request
	// to the SimpleWebHook Listener function. The WebhookQueryHandler should not receive the request and process it.

	// Create a channel to receive the message from the SimpleWebhookListener
	c := make(chan bool, 1)
	done := make(chan os.Signal, 1)
	dump := make(chan bool, 1)

	go server.CreateServer(&addr, server_port, done, webhook_tracker.NewLocalWebhookState())
	go simpleWebhookListener(c, target_listener)
	go simpleWebhookListener(dump, dump_listener)

	// Wait for 2 seconds to allow the servers to start
	time.Sleep(2 * time.Second)

	// Make test data
	raw := "Test"
	timeVal := time.Now()
	derived := make(map[string]interface{})
	derived["fieldname"] = 29
	var testEventData = go_system_api.EventData{
		Raw:       &raw,
		IndexTime: &timeVal,
		TimeStamp: &timeVal,
		EventType: &raw,
		Category:  &raw,
		Derived:   derived,
	}

	var testCommandList = go_system_api.CommandList{
		QueryId: "Test Command 1",
		Commands: []go_system_api.CommandStep{
			{CommandName: "Command 1", Args: "fieldname>=30 http://localhost:" + target_listener + "/webhook", Url: "http://localhost:" + server_port + "/query"},
			{CommandName: "Command 2", Args: "args 2", Url: "http://localhost:" + dump_listener + "/webhook"},
		},
		Step:     0,
		ErrorUrl: "Error Url",
	}

	var testProcessingEvent = go_system_api.ProcessingEvent{
		Commands: testCommandList,
		Event:    testEventData,
	}

	log.Println("Sending test event")
	// Send the request
	err := sendEvent(&testProcessingEvent, "http://localhost:"+server_port+"/query")
	if err != nil {
		log.Fatalf("Error sending test event: %s", err)
	}

	log.Println("Waiting for webhook to be received")

	// If channel C receives a message, the test fails, if it doesn't, the test passes
	select {
	case <-c:
		log.Fatalf("Test Failed")
		t.Fail()
	case <-time.After(5 * time.Second):
		log.Println("Test Passed")
		return
	}

}

func TestStringComparison(t *testing.T) {
	var server_port = "9105"
	var target_listener = "9106"
	var dump_listener = "9107"

	t.Log("Testing for a successful string comparison.")

	// This test sends a processing event with a string field rather than an integer field.
	// This function tests the WebhookQueryHandler, by creating an instance of it, then sending an example request
	// to the SimpleWebHook Listener function. The WebhookQueryHandler should receive the request and process it.

	// Create a channel to receive the message from the SimpleWebhookListener
	c := make(chan bool, 1)
	done := make(chan os.Signal, 1)
	dump := make(chan bool, 1)

	go server.CreateServer(&addr, server_port, done, webhook_tracker.NewLocalWebhookState())
	go simpleWebhookListener(c, target_listener)
	go simpleWebhookListener(dump, dump_listener)

	// Wait for 2 seconds to allow the servers to start
	time.Sleep(2 * time.Second)

	// Make test data
	raw := "Test"
	timeVal := time.Now()
	derived := make(map[string]interface{})
	derived["fieldname"] = "orange"
	var testEventData = go_system_api.EventData{
		Raw:       &raw,
		IndexTime: &timeVal,
		TimeStamp: &timeVal,
		EventType: &raw,
		Category:  &raw,
		Derived:   derived,
	}

	var testCommandList = go_system_api.CommandList{
		QueryId: "Test Command 1",
		Commands: []go_system_api.CommandStep{
			{CommandName: "Command 1", Args: "fieldname=orange http://localhost:" + target_listener + "/webhook", Url: "http://localhost:" + server_port + "/query"},
			{CommandName: "Command 2", Args: "args 2", Url: "http://localhost:" + dump_listener + "/webhook"},
		},
		Step:     0,
		ErrorUrl: "Error Url",
	}

	var testProcessingEvent = go_system_api.ProcessingEvent{
		Commands: testCommandList,
		Event:    testEventData,
	}

	log.Println("Sending test event")
	// Send the request
	err := sendEvent(&testProcessingEvent, "http://localhost:"+server_port+"/query")
	if err != nil {
		log.Fatalf("Error sending test event: %s", err)
	}

	log.Println("Waiting for webhook to be received")

	// Wait for the message from the SimpleWebhookListener
	select {
	case <-c:
		log.Println("Test Passed")
		return
	case <-time.After(5 * time.Second):
		log.Fatalf("Test Failed")
		t.FailNow()
	}

}

func sendEvent(event *go_system_api.ProcessingEvent, url string) error {
	// Send a Post request to the url
	jsonData, err := json.Marshal(event)
	if err != nil {
		log.Printf("Error marshalling event: %s \n", err)
		return err
	}

	for i := 1; i < 4; i++ {
		log.Printf("Sending event to: %s \n", url)
		r, err := http.Post(url, "application/json", strings.NewReader(string(jsonData)))
		if err == nil && r.StatusCode == 200 {
			return nil
		} else {
			log.Printf("Error forwarding event: %s \n", err)
		}
	}
	return err
}
