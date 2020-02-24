package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/getsentry/sentry-go"
	"github.com/gofrs/uuid"
	catalog "github.com/retgits/acme-serverless-catalog"
	"github.com/retgits/acme-serverless-catalog/internal/datastore/dynamodb"
)

func handleError(area string, headers map[string]string, err error) (events.APIGatewayProxyResponse, error) {
	sentry.CaptureException(fmt.Errorf("error %s: %s", area, err.Error()))
	msg := fmt.Sprintf("error %s: %s", area, err.Error())
	log.Println(msg)
	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusInternalServerError,
		Body:       msg,
		Headers:    headers,
	}, nil
}

func handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	sentrySyncTransport := sentry.NewHTTPSyncTransport()
	sentrySyncTransport.Timeout = time.Second * 3

	sentry.Init(sentry.ClientOptions{
		Dsn:         os.Getenv("SENTRY_DSN"),
		Transport:   sentrySyncTransport,
		ServerName:  os.Getenv("FUNCTION_NAME"),
		Release:     os.Getenv("VERSION"),
		Environment: os.Getenv("STAGE"),
	})

	headers := request.Headers
	if headers == nil {
		headers = make(map[string]string)
	}
	headers["Access-Control-Allow-Origin"] = "*"

	// Update the product with an ID
	prod, err := catalog.UnmarshalProduct(request.Body)
	if err != nil {
		return handleError("unmarshalling product", headers, err)
	}
	prod.ID = uuid.Must(uuid.NewV4()).String()

	dynamoStore := dynamodb.New()
	err = dynamoStore.AddProduct(prod)
	if err != nil {
		return handleError("adding product", headers, err)
	}

	status := catalog.ProductCreateResponse{
		Message:    "Product created successfully!",
		ResourceID: prod,
		Status:     http.StatusOK,
	}

	payload, err := status.Marshal()
	if err != nil {
		return handleError("marshalling response", headers, err)
	}

	response := events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Body:       payload,
		Headers:    headers,
	}

	return response, nil
}

// The main method is executed by AWS Lambda and points to the handler
func main() {
	lambda.Start(handler)
}
