// Package dynamodb leverages Amazon DynamoDB, a key-value and document database that delivers single-digit millisecond
// performance at any scale to store data.
package dynamodb

import (
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	catalog "github.com/retgits/acme-serverless-catalog"
	"github.com/retgits/acme-serverless-catalog/internal/datastore"
)

// The pointer to DynamoDB provides the API operation methods for making requests to Amazon DynamoDB.
// This specifically creates a single instance of the dynamoDB service which can be reused if the
// container stays warm.
var dbs *dynamodb.DynamoDB

// manager is an empty struct that implements the methods of the
// Manager interface.
type manager struct{}

// init creates the connection to dynamoDB. If the environment variable
// DYNAMO_URL is set, the connection is made to that URL instead of
// relying on the AWS SDK to provide the URL
func init() {
	awsSession := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(os.Getenv("REGION")),
	}))

	if len(os.Getenv("DYNAMO_URL")) > 0 {
		awsSession.Config.Endpoint = aws.String(os.Getenv("DYNAMO_URL"))
	}

	dbs = dynamodb.New(awsSession)
}

// New creates a new datastore manager using Amazon DynamoDB as backend
func New() datastore.Manager {
	return manager{}
}

// AddProduct stores a new product in Amazon DynamoDB
func (m manager) AddProduct(p catalog.Product) error {
	// Marshal the newly updated product struct
	payload, err := p.Marshal()
	if err != nil {
		return err
	}

	// Create a map of DynamoDB Attribute Values containing the table keys
	km := make(map[string]*dynamodb.AttributeValue)
	km["PK"] = &dynamodb.AttributeValue{
		S: aws.String("PRODUCT"),
	}
	km["SK"] = &dynamodb.AttributeValue{
		S: aws.String(p.ID),
	}

	// Create a map of DynamoDB Attribute Values containing the table data elements
	em := make(map[string]*dynamodb.AttributeValue)
	em[":payload"] = &dynamodb.AttributeValue{
		S: aws.String(string(payload)),
	}

	uii := &dynamodb.UpdateItemInput{
		TableName:                 aws.String(os.Getenv("TABLE")),
		Key:                       km,
		ExpressionAttributeValues: em,
		UpdateExpression:          aws.String("SET Payload = :payload"),
	}

	_, err = dbs.UpdateItem(uii)
	if err != nil {
		return err
	}

	return nil
}

// GetProduct retrieves a single product from DynamoDB based on the productID
func (m manager) GetProduct(productID string) (catalog.Product, error) {
	// Create a map of DynamoDB Attribute Values containing the table keys
	// for the access pattern PK = PRODUCT SK = ID
	km := make(map[string]*dynamodb.AttributeValue)
	km[":type"] = &dynamodb.AttributeValue{
		S: aws.String("PRODUCT"),
	}
	km[":id"] = &dynamodb.AttributeValue{
		S: aws.String(productID),
	}

	// Create the QueryInput
	qi := &dynamodb.QueryInput{
		TableName:                 aws.String(os.Getenv("TABLE")),
		KeyConditionExpression:    aws.String("PK = :type AND SK = :id"),
		ExpressionAttributeValues: km,
	}

	// Execute the DynamoDB query
	qo, err := dbs.Query(qi)
	if err != nil {
		return catalog.Product{}, err
	}

	// Return an error if no product was found
	if len(qo.Items) == 0 {
		return catalog.Product{}, fmt.Errorf("Unable to find product with id %s", productID)
	}

	// Create a product struct from the data
	str := *qo.Items[0]["Payload"].S
	return catalog.UnmarshalProduct(str)
}

// GetProducts retrieves all products from DynamoDB
func (m manager) GetProducts() ([]catalog.Product, error) {
	// Create a map of DynamoDB Attribute Values containing the table keys
	// for the access pattern PK = PRODUCT
	km := make(map[string]*dynamodb.AttributeValue)
	km[":type"] = &dynamodb.AttributeValue{
		S: aws.String("PRODUCT"),
	}

	// Create the QueryInput
	qi := &dynamodb.QueryInput{
		TableName:                 aws.String(os.Getenv("TABLE")),
		KeyConditionExpression:    aws.String("PK = :type"),
		ExpressionAttributeValues: km,
	}

	qo, err := dbs.Query(qi)
	if err != nil {
		return nil, err
	}

	prods := make([]catalog.Product, len(qo.Items))

	for idx, ct := range qo.Items {
		str := *ct["Payload"].S
		prod, err := catalog.UnmarshalProduct(str)
		if err != nil {
			log.Println(fmt.Sprintf("error unmarshalling product data: %s", err.Error()))
			continue
		}
		prods[idx] = prod
	}

	return prods, nil
}
