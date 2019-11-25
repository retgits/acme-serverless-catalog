package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/gofrs/uuid"
	"github.com/kelseyhightower/envconfig"
	"github.com/retgits/catalog"
	wflambda "github.com/retgits/wavefront-lambda-go"
)

var wfAgent = wflambda.NewWavefrontAgent(&wflambda.WavefrontConfig{})

// config is the struct that is used to keep track of all environment variables
type config struct {
	AWSRegion     string `required:"true" split_words:"true" envconfig:"REGION"`
	DynamoDBTable string `required:"true" split_words:"true" envconfig:"TABLENAME"`
}

var c config

func logError(stage string, err error) (events.APIGatewayProxyResponse, error) {
	errormessage := fmt.Sprintf("error %s: %s", stage, err.Error())
	log.Println(errormessage)
	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusInternalServerError,
		Body:       errormessage,
	}, err
}

func handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	response := events.APIGatewayProxyResponse{}

	// Get configuration set using environment variables
	err := envconfig.Process("", &c)
	if err != nil {
		return logError("starting function", err)
	}

	// Create an AWS session
	awsSession := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(c.AWSRegion),
	}))

	// Create a DynamoDB session
	dbs := dynamodb.New(awsSession)

	// Update the product with an ID
	prod, err := catalog.UnmarshalProduct(request.Body)
	if err != nil {
		return logError("unmarshalling product", err)
	}
	prod.ID = uuid.Must(uuid.NewV4()).String()

	// Marshal the newly updated product struct
	payload, err := prod.Marshal()
	if err != nil {
		return logError("marshalling product", err)
	}

	// Create a map of DynamoDB Attribute Values containing the table keys
	km := make(map[string]*dynamodb.AttributeValue)
	km["ID"] = &dynamodb.AttributeValue{
		S: aws.String(prod.ID),
	}

	em := make(map[string]*dynamodb.AttributeValue)
	em[":content"] = &dynamodb.AttributeValue{
		S: aws.String(payload),
	}

	uii := &dynamodb.UpdateItemInput{
		TableName:                 aws.String(c.DynamoDBTable),
		Key:                       km,
		ExpressionAttributeValues: em,
		UpdateExpression:          aws.String("SET ProductContent = :content"),
	}

	_, err = dbs.UpdateItem(uii)
	if err != nil {
		return logError("updating dynamodb", err)
	}

	status := catalog.ProductCreateResponse{
		Message:    "Product created successfully!",
		ResourceID: prod,
		Status:     http.StatusOK,
	}

	statusPayload, err := status.Marshal()
	if err != nil {
		return logError("marshalling response", err)
	}

	response.StatusCode = http.StatusOK
	response.Body = statusPayload

	return response, nil
}

// The main method is executed by AWS Lambda and points to the handler
func main() {
	lambda.Start(wfAgent.WrapHandler(handler))
}
