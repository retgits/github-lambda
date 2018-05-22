/*
Package main is the main executable of the serverless function. It will query the GitHub
API and search for issues assigned to the user whose Personal Access Token is used. For
each issue an event will be sent to a Trello function to create a new Trello card
*/
package main

// The imports
import (
	"context"
	"encoding/json"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/aws/aws-lambda-go/events"
	rt "github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/aws/aws-xray-sdk-go/xray"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

// Variables that are set as Environment Variables
var (
	githubAppToken     = os.Getenv("apptoken")
	githubTimeInterval = os.Getenv("interval")
	trelloARN          = os.Getenv("arntrello")
	region             = "us-west-2"
)

type lambdaEvent struct {
	EventVersion string
	EventSource  string
	Trello       trelloEvent
}

type trelloEvent struct {
	Title       string
	Description string
}

// The handler function is executed every time that a new Lambda event is received.
// It takes a JSON payload (you can see an example in the event.json file) and only
// returns an error if the something went wrong. The event comes fom CloudWatch and
// is scheduled every interval (where the interval is defined as variable)
func handler(request events.CloudWatchEvent) error {
	// stdout and stderr are sent to AWS CloudWatch Logs
	log.Printf("Processing Lambda request [%s]", request.ID)

	// Create a new GitHub client
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: githubAppToken},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	// Create a new time to make sure we check for new issues from the previous
	// execution of this function.
	i, _ := strconv.Atoi(githubTimeInterval)
	interval := time.Duration(i) * time.Minute
	t := request.Time.Add(-interval)
	log.Printf("Check GitHub issues for the current user since [%s]", t)

	// Get all the issues assigned to the current authenticated user
	issueOpts := github.IssueListOptions{Since: t}
	issues, _, err := client.Issues.List(ctx, false, &issueOpts)

	if err != nil {
		log.Print(err)
		return err
	}

	if len(issues) == 0 {
		log.Printf("There are no new issues")
	}

	// Create a new AWS session to invoke a Lambda function
	config := aws.NewConfig().WithRegion(region)
	aws := lambda.New(session.New(config))
	xray.AWS(aws.Client)

	// For each new issue create a Trello card
	for _, issue := range issues {
		payload := lambdaEvent{
			EventVersion: "1.0",
			EventSource:  "aws:lambda",
			Trello: trelloEvent{
				Title:       issue.GetTitle(),
				Description: "Repository: " + issue.GetRepository().GetHTMLURL() + "\nDirect link: " + issue.GetHTMLURL(),
			},
		}

		var b []byte
		b, _ = json.Marshal(payload)

		// Execute the call to the Trello Lambda function
		_, err := aws.InvokeWithContext(ctx, &lambda.InvokeInput{
			FunctionName: &trelloARN,
			Payload:      b})

		if err != nil {
			log.Printf(err.Error())
			return err
		}

		log.Printf("Created a card for %s\n", issue.GetTitle())
	}
	return nil
}

// The main method is executed by AWS Lambda and points to the handler
func main() {
	rt.Start(handler)
}
