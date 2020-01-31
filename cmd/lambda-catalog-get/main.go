package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/retgits/acme-serverless-catalog/internal/datastore/dynamodb"
)

func handleError(area string, err error) (events.APIGatewayProxyResponse, error) {
	msg := fmt.Sprintf("error %s: %s", area, err.Error())
	log.Println(msg)
	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusInternalServerError,
		Body:       msg,
	}, err
}

func handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	response := events.APIGatewayProxyResponse{}

	// Create the key attributes
	productID := request.PathParameters["id"]

	dynamoStore := dynamodb.New()
	prod, err := dynamoStore.GetProduct(productID)
	if err != nil {
		return handleError("finding product", err)
	}

	statusPayload, err := prod.Marshal()
	if err != nil {
		return handleError("marshalling response", err)
	}

	headers := request.Headers
	headers["Access-Control-Allow-Origin"] = "*"

	response.StatusCode = http.StatusOK
	response.Body = statusPayload
	response.Headers = headers

	return response, nil
}

// The main method is executed by AWS Lambda and points to the handler
func main() {
	lambda.Start(handler)
}
