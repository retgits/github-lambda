//go:generate go run $GOPATH/src/github.com/TIBCOSoftware/flogo-lib/flogo/gen/gen.go $GOPATH

// Package main is the main executable of the serverless function. It will query the GitHub
// API and search for issues assigned to the user whose Personal Access Token is used. For
// each issue an event will be sent to a Trello function to create a new Trello card
package main

// The ever important imports
import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	lmb "github.com/TIBCOSoftware/flogo-contrib/activity/lambda"
	"github.com/TIBCOSoftware/flogo-contrib/trigger/lambda"
	"github.com/TIBCOSoftware/flogo-lib/config"
	"github.com/TIBCOSoftware/flogo-lib/core/data"
	"github.com/TIBCOSoftware/flogo-lib/flogo"
	"github.com/TIBCOSoftware/flogo-lib/logger"
	"github.com/google/go-github/github"
	"github.com/retgits/flogo-components/activity/awsssm"
	"github.com/retgits/flogo-components/activity/envkey"
	"github.com/retgits/flogo-components/activity/githubissues"
)

// Constants
const (
	// The name of the GitHub Personal Access Token parameter in Amazon SSM
	accessToken = "/github/apptoken"
	// The name of the time interval parameter in Amazon SSM
	interval = "/github/interval"
	// The name of the Trello ARN parameter in Amazon SSM
	trelloARN = "/trello/arn"
	// The default region
	defaultRegion = "us-west-2"
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

// Init makes sure that everything is ready to go!
func init() {
	config.SetDefaultLogLevel("INFO")
	logger.SetLogLevel(logger.InfoLevel)

	app := shimApp()

	e, err := flogo.NewEngine(app)

	if err != nil {
		logger.Error(err)
		return
	}

	e.Init(true)
}

// shimApp is used to build a new Flogo app and register the Lambda trigger with the engine.
// The shimapp is used by the shim, which triggers the engine every time an event comes into Lambda.
func shimApp() *flogo.App {
	// Create a new Flogo app
	app := flogo.NewApp()

	// Register the Lambda trigger with the Flogo app
	trg := app.NewTrigger(&lambda.LambdaTrigger{}, nil)
	trg.NewFuncHandler(nil, RunActivities)

	// Return a pointer to the app
	return app
}

// RunActivities is where the magic happens. This is where you get the input from any event that might trigger
// your Lambda function in a map called evt (which is part of the inputs). The below sample,
// will simply log "Go Serverless v1.x! Your function executed successfully!" and return the same as a response.
// The trigger, in main.go, will take care of marshalling it into a proper response for the API Gateway
func RunActivities(ctx context.Context, inputs map[string]*data.Attribute) (map[string]*data.Attribute, error) {
	// Get the items from SSM
	in := map[string]interface{}{"action": "retrieveList", "parameterName": fmt.Sprintf("%s,%s,%s", accessToken, interval, trelloARN), "decryptParameter": true}
	out, err := flogo.EvalActivity(&awsssm.MyActivity{}, in)
	if err != nil {
		return nil, err
	}

	ghToken := out["result"].Value().(map[string]interface{})[accessToken].(string)
	arn := out["result"].Value().(map[string]interface{})[trelloARN].(string)

	timeInterval, err := strconv.Atoi(out["result"].Value().(map[string]interface{})[interval].(string))
	if err != nil {
		return nil, err
	}

	// Get the region as environment variable
	in = map[string]interface{}{"envkey": "region", "fallback": defaultRegion}
	out, err = flogo.EvalActivity(&envkey.MyActivity{}, in)
	if err != nil {
		return nil, err
	}
	region := out["result"].Value().(string)

	// Get GitHub issues
	in = map[string]interface{}{"token": ghToken, "timeInterval": timeInterval}
	out, err = flogo.EvalActivity(&githubissues.MyActivity{}, in)
	if err != nil {
		return nil, err
	}
	ghIssues := out["result"].Value().([]interface{})

	// Send issues to Trello
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

		in = map[string]interface{}{"arn": arn, "region": region, "payload": payloadMap}
		_, err := flogo.EvalActivity(&lmb.Activity{}, in)
		if err != nil {
			return nil, err
		}
	}

	logger.Infof("Sent %d issues to Trello", len(ghIssues))

	return nil, nil
}
