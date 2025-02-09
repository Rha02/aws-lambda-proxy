package main

import (
	"encoding/json"
	"fmt"
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
		errMsg, _ := json.Marshal(err.Error())
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			Body: fmt.Sprintf(`{
				"error": "Failed to send request to the url",
				"details": %s
			}`, errMsg),
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

	// Get response content-type
	contentType := res.Header.Get("Content-Type")

	// Return response
	return events.APIGatewayProxyResponse{
		StatusCode: res.StatusCode,
		Headers: map[string]string{
			"Content-Type": contentType,
		},
		Body: string(resBody),
	}, nil
}

func main() {
	lambda.Start(requestHandler)
}
