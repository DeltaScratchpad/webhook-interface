package processing

import (
	"fmt"
	"github.com/DeltaScratchpad/go-system-api"
	"github.com/DeltaScratchpad/webhook-interface/helpers"
	"github.com/DeltaScratchpad/webhook-interface/webhook-tracker"
	"regexp"
	"strconv"
)

func ProcessProcessingEvent(query *go_system_api.ProcessingEvent, state webhook_tracker.WebhookState) {

	//Parse args
	field, relation, threshold_int, threshold_str, webhook, err, isInt := parseArgs(&query.Commands.Commands[query.Commands.Step].Args)
	if err != nil {
		helpers.LogError(fmt.Sprintf("Failed to parse args: %s", err), query)
		return
	}

	var result bool = false
	// Get field value
	value, err := helpers.GetIntValue(&query.Event, field)
	if err == nil && isInt {
		result = compareIntByRelation(value, threshold_int, relation)
		//log.Printf("Result: %t, Value: %d\n", result, value)
	} else {
		value, err := helpers.GetStringValue(&query.Event, field)
		if err == nil {
			if relation == "=" {
				result = value == threshold_str
				//log.Printf("Result: %t, Value: %s\n", result, value)
			}
		} else {
			// We didn't find the field, so we can't compare it.
			return
		}
	}

	// Check if we should call it, but only if it hasn't already been called.
	// Under a race condition, multiple could be sent.
	//TODO: Would need distributed locking to resolve.
	if result && !state.HasBeenCalled(webhook, query.Commands.QueryId) {
		err := helpers.SendGetWebhook(webhook)
		if err != nil {
			state.IncrementCallCount(webhook, query.Commands.QueryId)
			helpers.LogError(fmt.Sprintf("Failed to send webhook: %s", err), query)
			return
		}
	}
}

var args_parser = regexp.MustCompile(`(?P<field>\w+)(?P<relation>[><=]{1,2})(?P<threshold>[\d\w]+)\s(?P<webhook>.+)`)
var filed_index = args_parser.SubexpIndex(`field`)
var relation_index = args_parser.SubexpIndex(`relation`)
var threshold_index = args_parser.SubexpIndex(`threshold`)
var webhook_index = args_parser.SubexpIndex(`webhook`)

func parseArgs(args *string) (field string, relation string, threshold_int int, threshold_str string, webhook string, err error, isInt bool) {
	matches := args_parser.FindStringSubmatch(*args)
	var int_err error

	field = matches[filed_index]
	relation = matches[relation_index]
	threshold_str = matches[threshold_index]
	threshold_int, int_err = strconv.Atoi(threshold_str)
	webhook = matches[webhook_index]
	if field == "" || relation == "" || threshold_str == "" || webhook == "" {
		err = fmt.Errorf("invalid args")
	}
	if int_err == nil {
		isInt = true
	} else {
		isInt = false
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
