package main

import (
	"fmt"

	"github.com/aws/aws-lambda-go/events"
)

func lambdaErrResponse(httpStatus int, errMsg string) events.APIGatewayProxyResponse {
	body := fmt.Sprintf(`{"error": "%s"}`, errMsg)

	return lambdaHttpResponse(httpStatus, map[string]string{
		"Content-Type": "application/json",
	}, body)
}

func lambdaHttpResponse(httpStatus int, headers map[string]string, body string) events.APIGatewayProxyResponse {
	return events.APIGatewayProxyResponse{
		StatusCode: httpStatus,
		Headers:    headers,
		Body:       body,
	}
}
