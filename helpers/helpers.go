package helpers

import (
	"bytes"
	"fmt"
	go_system_api "github.com/DeltaScratchpad/go-system-api"

	// jsoniter is used for increased performance.
	// The standard library would also work, if dependencies are not permitted.
	jsoniter "github.com/json-iterator/go"

	"log"
	"net/http"
	"strconv"
)

func LogError(err string, event *go_system_api.ProcessingEvent) {
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	var err_url = event.Commands.ErrorUrl
	if err_url != "" {
		var errorBody = go_system_api.ErrorBody{
			QueryID:     event.Commands.QueryId,
			Step:        int64(event.Commands.Step),
			Recoverable: true,
			ErrorMsg:    err,
			DebugMsg:    err,
			Event:       &event.Event,
		}
		jsonData, err := json.Marshal(errorBody)
		if err != nil {
			log.Printf("Error marshalling error body: %s \n", err)
			return
		}
		for i := 1; i < 4; i++ {
			r, err := http.Post(err_url, "application/json", bytes.NewBuffer(jsonData))
			if err == nil && r.StatusCode == 200 {
				return
			} else {
				log.Fatalf("Error logging error: %s \n", err)
			}
		}
	} else {
		log.Printf("Error URL was nil for event. Won't be able to log errors.")
	}
}

func ForwardEvent(event *go_system_api.ProcessingEvent) {
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	event.Commands.Step += 1

	jsonData, err := json.Marshal(event)
	if err != nil {
		log.Printf("Error marshalling event: %s \n", err)
		LogError(fmt.Sprintf("Failed to serialise event: %s", err), event)
		return
	}

	for i := 1; i < 4; i++ {
		r, err := http.Post(event.Commands.Commands[event.Commands.Step].Url, "application/json", bytes.NewBuffer(jsonData))
		if err == nil && r.StatusCode == 200 {
			return
		} else {
			log.Fatalf("Error forwarding event: %s \n", err)
		}
	}

}

func ParseProcessingEvent(w http.ResponseWriter, r *http.Request) (go_system_api.ProcessingEvent, error) {
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	var query go_system_api.ProcessingEvent
	if err := json.NewDecoder(r.Body).Decode(&query); err != nil {
		log.Printf("Error decoding query: %s \n", err)
		InternalServerErrorHandler(w, r)
		//TODO! Forward event
		return go_system_api.ProcessingEvent{}, nil
	}
	return query, nil
}

func InternalServerErrorHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte("Internal Server Error"))
	return
}

func InvalidHttpMethodHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusMethodNotAllowed)
	w.Write([]byte("Method Not Allowed"))
	return
}

func GetStringValue(event *go_system_api.EventData, field string) (string, error) {
	switch field {
	case "raw":
		if event.Raw != nil {
			return *event.Raw, nil
		} else {
			return "", fmt.Errorf("field %s was nil", field)
		}
	case "event_type":
		if event.EventType != nil {
			return *event.EventType, nil
		} else {
			return "", fmt.Errorf("field %s was nil", field)
		}
	case "category":
		if event.Category != nil {
			return *event.Category, nil
		} else {
			return "", fmt.Errorf("field %s was nil", field)
		}
	default:
		if value, ok := event.Derived[field]; ok {
			return value.(string), nil
		} else {
			return "", fmt.Errorf("field %s was nil", field)
		}
	}
}

func GetIntValue(event *go_system_api.EventData, field string) (int, error) {
	switch field {
	case "raw":
		if event.Raw != nil {
			return strconv.Atoi(*event.Raw)
		} else {
			return 0, fmt.Errorf("field %s was nil", field)
		}
	case "event_type":
		if event.EventType != nil {
			return strconv.Atoi(*event.EventType)
		} else {
			return 0, fmt.Errorf("field %s was nil", field)
		}
	case "category":
		if event.Category != nil {
			return strconv.Atoi(*event.Category)
		} else {
			return 0, fmt.Errorf("field %s was nil", field)
		}
	default:
		if value, ok := event.Derived[field]; ok {
			switch value.(type) {
			case int:
				return value.(int), nil
			case float64:
				return int(value.(float64)), nil
			case string:
				return strconv.Atoi(value.(string))
			default:
				return 0, fmt.Errorf("field %s was incompatible type", field)
			}
		} else {
			return 0, fmt.Errorf("field %s was nil", field)
		}
	}

}

func SendGetWebhook(webhook string) (err error) {
	var res *http.Response
	for i := 0; i < 4; i++ {
		res, err = http.Get(webhook)
		if err == nil && res.StatusCode == 200 {
			return
		}
	}
	return
}
