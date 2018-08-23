# github-lambda - A serverless app to get GitHub issues

This serverless function is designed to query the Github API for new issues assigned to the current user.

## Layout
```bash
.                    
├── test            
│   └── event.json      <-- Sample event to test using SAM local
├── .gitignore          <-- Ignoring the things you do not want in git
├── function.go         <-- Test main function code
├── LICENSE             <-- The license file
├── main.go             <-- The Flogo Lambda trigger code
├── Makefile            <-- Makefile to build and deploy
├── README.md           <-- This file
└── template.yaml       <-- SAM Template
```

## Installing
There are a few ways to install this project

### Get the sources
You can get the sources for this project by simply running
```bash
$ go get -u github.com/retgits/github-lambda/...
```

### Deploy
Deploy the Lambda app by running
```bash
$ make deploy
```

## Parameters
### AWS Systems Manager parameters
The code will automatically retrieve the below list of parameters from the AWS Systems Manager Parameter store:

* **/github/apptoken**: Your Personal Access Token from GitHub
* **/github/interval**: The timeinterval in which this function is triggered
* **/trello/arn**: The ARN for the Trello function

### Deployment parameters
In the `template.yaml` there are certain deployment parameters:

* **region**: The AWS region in which the code is deployed

## Make targets
github-lambda has a _Makefile_ that can be used for most of the operations

```
usage: make [target]
```

* **deps**: Gets all dependencies for this app
* **clean** : Removes the dist directory
* **build**: Builds an executable to be deployed to AWS Lambda
* **test-lambda**: Clean, builds and tests the code by using the AWS SAM CLI
* **deploy**: Cleans, builds and deploys the code to AWS Lambda