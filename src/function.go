//go:generate go run ../../../TIBCOSoftware/flogo-lib/flogo/gen/gen.go $GOPATH

/*
Package main is the main executable of the serverless function. It will query the GitHub
API and search for issues assigned to the user whose Personal Access Token is used. For
each issue an event will be sent to a Trello function to create a new Trello card
*/
package main

// The imports
import (
	"encoding/json"
	"os"
	"strconv"

	"github.com/TIBCOSoftware/flogo-contrib/activity/lambda"
	"github.com/TIBCOSoftware/flogo-lib/flogo"
	"github.com/TIBCOSoftware/flogo-lib/logger"
	"github.com/aws/aws-lambda-go/events"
	"github.com/google/go-github/github"
	"github.com/retgits/flogo-components/activity/githubissues"
)

var (
	// Your Personal Access Token from GitHub
	accessToken = os.Getenv("apptoken")

	// The timeinterval in which this function is triggered
	interval = os.Getenv("interval")

	// The ARN for the Trello function
	trelloARN = os.Getenv("P_trello_arn")

	// The region in which the Trello function runs
	region = "us-west-2"
)

// LambdaEvent is the outer structure of the events that are received by this function
type LambdaEvent struct {
	EventVersion string
	EventSource  string
	Event        interface{}
}

// TrelloEvent is the structure for the data representing a TrelloCard
type TrelloEvent struct {
	Title       string
	Description string
}

// Invoke is executed every time a new Lambda event is received.
// It takes the payload event (you can see an example in the event.json file) and
// returns a map[string]interface{} representing a JSON payload and an optional error
func Invoke(request events.CloudWatchEvent) (map[string]interface{}, error) {
	logger.Infof("Parsing interval to integer")
	timeInterval, err := strconv.Atoi(interval)
	if err != nil {
		return nil, err
	}

	logger.Infof("Getting issues from GitHub")
	in := map[string]interface{}{"token": accessToken, "timeInterval": timeInterval}
	out, err := flogo.EvalActivity(&githubissues.MyActivity{}, in)
	if err != nil {
		return nil, err
	}
	ghIssues := out["result"].Value().([]interface{})

	logger.Infof("Send issues to Trello")
	for _, val := range ghIssues {
		issue := val.(*github.Issue)

		trelloEvent := TrelloEvent{
			Description: "Repository: " + issue.GetRepository().GetHTMLURL() + "\nDirect link: " + issue.GetHTMLURL(),
			Title:       issue.GetTitle(),
		}
		payload := LambdaEvent{
			EventVersion: "1.0",
			EventSource:  "aws:lambda",
			Event:        trelloEvent,
		}

		var payloadMap map[string]interface{}
		inrec, _ := json.Marshal(payload)
		json.Unmarshal(inrec, &payloadMap)

		in = map[string]interface{}{"arn": trelloARN, "region": region, "payload": payloadMap}
		_, err := flogo.EvalActivity(&lambda.Activity{}, in)
		if err != nil {
			return nil, err
		}
	}

	logger.Infof("Sent all issues to Trello")

	return map[string]interface{}{"data": "OK", "status": 200}, nil
}
