package main

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

type RequestBody struct {
	Url string `json:"url"`
}

func requestHandler(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Parse request body
	var body RequestBody
	if err := json.Unmarshal([]byte(req.Body), &body); err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			Body: `{"error": "Failed to parse request JSON body"}`,
		}, nil
	}

	// Send Proxy request
	res, err := http.Get(body.Url)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			Body: `{"error": "Failed to send request to the url"}`,
		}, nil
	}
	// Defer closing response body
	defer res.Body.Close()

	// Read response body
	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			Body: `{"error": "Failed to read response JSON body"}`,
		}, nil
	}

	// Return response
	return events.APIGatewayProxyResponse{
		StatusCode: res.StatusCode,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: string(resBody),
	}, nil
}

func main() {
	lambda.Start(requestHandler)
}
