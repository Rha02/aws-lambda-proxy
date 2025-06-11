## Description
This is a simple proxy built utilizing AWS Lambda functions.

## Compiling
On a Linux machine, run the following commands:
```sh
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -tags lambda.norpc -o bootstrap *.go
```
Zip the compiled binary:
```
zip bootstrap.zip bootstrap
```
Upload the zip file to AWS Lambda.

## Documentation
The proxy function currently can only make a simple HTTP GET request.

### Request Body breakdown
- `url` - the url to which the proxy must send request to.


### Example Usage

```sh
curl --location '<aws-lambda-function-url>' \
--header 'Content-Type: application/json' \
--data '{
    "url": "https://ipinfo.io/json"
}'
```
