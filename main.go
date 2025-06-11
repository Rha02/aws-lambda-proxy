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
		return lambdaErrResponse(400, "Invalid request body format"), nil
	}

	if proxyReq.URL == "" {
		return lambdaErrResponse(400, "Missing url in request body"), nil
	}

	if _, err := url.Parse(proxyReq.URL); err != nil {
		return lambdaErrResponse(400, "Invalid URL format"), nil
	}

	if proxyReq.Method == "" {
		proxyReq.Method = "POST"
	}

	request, err := http.NewRequest(proxyReq.Method, proxyReq.URL, strings.NewReader(proxyReq.Body))
	if err != nil {
		return lambdaErrResponse(500, "Error creating request"), nil
	}

	for headerKey, headerValue := range proxyReq.Headers {
		request.Header.Set(headerKey, headerValue)
	}

	client := &http.Client{}
	res, err := client.Do(request)
	if err != nil {
		msg := fmt.Sprintf(`{"error": "Failed to send request to the url", "details": %s}`, err.Error())
		return lambdaErrResponse(500, msg), nil
	}
	defer res.Body.Close()

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return lambdaErrResponse(500, "Failed to read response body"), nil
	}

	resHeaders := make(map[string]string)
	for key := range res.Header {
		resHeaders[key] = res.Header.Get(key)
	}

	return lambdaHttpResponse(res.StatusCode, resHeaders, string(resBody)), nil
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
