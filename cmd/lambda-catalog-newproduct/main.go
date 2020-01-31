package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/gofrs/uuid"
	catalog "github.com/retgits/acme-serverless-catalog"
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

	// Update the product with an ID
	prod, err := catalog.UnmarshalProduct(request.Body)
	if err != nil {
		return handleError("unmarshalling product", err)
	}
	prod.ID = uuid.Must(uuid.NewV4()).String()

	dynamoStore := dynamodb.New()
	err = dynamoStore.AddProduct(prod)
	if err != nil {
		return handleError("adding product", err)
	}

	status := catalog.ProductCreateResponse{
		Message:    "Product created successfully!",
		ResourceID: prod,
		Status:     http.StatusOK,
	}

	statusPayload, err := status.Marshal()
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
