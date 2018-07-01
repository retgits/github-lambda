//go:generate go run ../../../TIBCOSoftware/flogo-lib/flogo/gen/gen.go $GOPATH

/*
Package main is the main executable of the serverless function. It will query the GitHub
API and search for issues assigned to the user whose Personal Access Token is used. For
each issue an event will be sent to a Trello function to create a new Trello card
*/
package main

// The imports
import (
	"github.com/TIBCOSoftware/flogo-lib/logger"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

// The handler function is executed every time that a new Lambda event is received.
// It takes a JSON payload (you can see an example in the event.json file) and only
// returns an error if the something went wrong. The event comes fom CloudWatch and
// is scheduled every interval (where the interval is defined as variable)
func handler(request events.CloudWatchEvent) error {
	// stdout and stderr are sent to AWS CloudWatch Logs
	logger.Infof("Processing Lambda request [%s]", request.ID)

	_, err := Invoke(request)
	if err != nil {
		logger.Infof("Error while sending data: %s", err.Error())
		return err
	}

	// Return no error
	return nil
}

// The main method is executed by AWS Lambda and points to the handler
func main() {
	lambda.Start(handler)
}
