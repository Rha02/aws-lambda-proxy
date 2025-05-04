package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/joho/godotenv"
)

func requestHandler(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	urlParams := req.QueryStringParameters
	target, ok := urlParams["url"]
	if !ok {
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			Body: `{"error": "Missing query parameter: url."}`,
		}, nil
	}

	request, err := http.NewRequest(req.HTTPMethod, target, strings.NewReader(req.Body))
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			Body: `{"error": "Error creating request"}`,
		}, nil
	}

	for headerKey, headerValue := range req.Headers {
		request.Header.Set(headerKey, headerValue)
	}

	client := &http.Client{}
	res, err := client.Do(request)
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
	defer res.Body.Close()

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

	resHeaders := make(map[string]string)
	for key, _ := range res.Header {
		resHeaders[key] = res.Header.Get(key)
	}

	return events.APIGatewayProxyResponse{
		StatusCode: res.StatusCode,
		Headers:    resHeaders,
		Body:       string(resBody),
	}, nil
}

func devToLambdaHandler(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)

	queryParamsMap := make(map[string]string)
	queryParams := r.URL.Query()
	for key, v := range queryParams {
		queryParamsMap[key] = v[0]
	}

	headers := make(map[string]string)
	for key, v := range r.Header {
		headers[key] = v[0]
	}

	lambdaReq := events.APIGatewayProxyRequest{
		Headers:               headers,
		HTTPMethod:            r.Method,
		QueryStringParameters: queryParamsMap,
		Body:                  string(body),
	}

	res, err := requestHandler(lambdaReq)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	w.WriteHeader(res.StatusCode)
	w.Write([]byte(res.Body))
}

func main() {
	godotenv.Load()

	environment := os.Getenv("ENVIRONMENT")
	if environment == "dev" {
		http.HandleFunc("/lambda", devToLambdaHandler)
		http.ListenAndServe(":8081", nil)
	} else {
		lambda.Start(requestHandler)
	}
}
