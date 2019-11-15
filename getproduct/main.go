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
	"github.com/kelseyhightower/envconfig"
	"github.com/retgits/catalog"
	wflambda "github.com/wavefronthq/wavefront-lambda-go"
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

	// Create the key attributes
	productID := request.PathParameters["id"]

	// Create a map of DynamoDB Attribute Values containing the table keys
	km := make(map[string]*dynamodb.AttributeValue)
	km[":productid"] = &dynamodb.AttributeValue{
		S: aws.String(productID),
	}

	si := &dynamodb.ScanInput{
		TableName:                 aws.String(c.DynamoDBTable),
		ExpressionAttributeValues: km,
		FilterExpression:          aws.String("UserID = :userid"),
	}

	so, err := dbs.Scan(si)
	if err != nil {
		return logError("scanning dynamodb", err)
	}

	if len(so.Items) == 0 {
		if err != nil {
			return logError("retrieving product data", fmt.Errorf("no product found with id %s", productID))
		}
	}

	str := *so.Items[0]["ProductContent"].S
	prod, err := catalog.UnmarshalProduct(str)
	if err != nil {
		errormessage := fmt.Sprintf("error unmarshalling product data: %s", err.Error())
		log.Println(errormessage)
	}

	statusPayload, err := prod.Marshal()
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
