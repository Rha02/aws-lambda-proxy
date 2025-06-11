package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/joho/godotenv"
)

type ProxyRequest struct {
	URL     string            `json:"url"`
	Method  string            `json:"method"`
	Headers map[string]string `json:"headers,omitempty"`
	Body    string            `json:"body,omitempty"`
}

func requestHandler(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var proxyReq ProxyRequest
	if err := json.Unmarshal([]byte(req.Body), &proxyReq); err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			Body: `{"error": "Invalid request body format"}`,
		}, nil
	}

	if proxyReq.URL == "" {
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			Body: `{"error": "Missing url in request body"}`,
		}, nil
	}

	if _, err := url.Parse(proxyReq.URL); err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			Body: `{"error": "Invalid URL format"}`,
		}, nil
	}

	if proxyReq.Method == "" {
		proxyReq.Method = "POST"
	}

	request, err := http.NewRequest(proxyReq.Method, proxyReq.URL, strings.NewReader(proxyReq.Body))
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			Body: `{"error": "Error creating request"}`,
		}, nil
	}

	for headerKey, headerValue := range proxyReq.Headers {
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
			Body: `{"error": "Failed to read response body"}`,
		}, nil
	}

	resHeaders := make(map[string]string)
	for key := range res.Header {
		resHeaders[key] = res.Header.Get(key)
	}

	return events.APIGatewayProxyResponse{
		StatusCode: res.StatusCode,
		Headers:    resHeaders,
		Body:       string(resBody),
	}, nil
}

func devToLambdaHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "Failed to read request body"}`))
		return
	}

	lambdaReq := events.APIGatewayProxyRequest{
		Body: string(body),
	}

	res, err := requestHandler(lambdaReq)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	for key, v := range res.Headers {
		w.Header().Add(key, v)
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
