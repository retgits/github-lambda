# GitHub app for Lambda

This serverless function is designed to query the Github API for new issues assigned to the current user.

## Layout
```bash
.
├── .travis.yml                 <-- Travis-CI build file
├── event.json                  <-- Sample event to test using SAM local
├── README.md                   <-- This file
├── src                         <-- Source code for a lambda function
│   ├── main.go                 <-- Lambda trigger code
│   └── function.go             <-- Lambda function code
└── template.yaml               <-- SAM Template
```

## Build and Deploy
Building and deploying this function is done through Travis-CI using [lambda-builder](https://github.com/retgits/lambda-builder)

## AWS Systems Manager
Within the AWS Systems Manager Parameter store there are three parameters that are used in this app:

* /github/apptoken
* /github/interval
* /trello/arn

## TODO
- [ ] Replace the trigger code with a proper Flogo trigger
