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

type manager struct{}

func New() datastore.Manager {
	return manager{}
}

func (m manager) AddProduct(p catalog.Product) error {
	awsSession := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(os.Getenv("REGION")),
	}))

	dbs := dynamodb.New(awsSession)

	// Marshal the newly updated product struct
	payload, err := p.Marshal()
	if err != nil {
		return err
	}

	// Create a map of DynamoDB Attribute Values containing the table keys
	km := make(map[string]*dynamodb.AttributeValue)
	km["ID"] = &dynamodb.AttributeValue{
		S: aws.String(p.ID),
	}

	em := make(map[string]*dynamodb.AttributeValue)
	em[":content"] = &dynamodb.AttributeValue{
		S: aws.String(payload),
	}

	uii := &dynamodb.UpdateItemInput{
		TableName:                 aws.String(os.Getenv("TABLE")),
		Key:                       km,
		ExpressionAttributeValues: em,
		UpdateExpression:          aws.String("SET ProductContent = :content"),
	}

	_, err = dbs.UpdateItem(uii)
	if err != nil {
		return err
	}

	return nil
}

func (m manager) GetProduct(productID string) (catalog.Product, error) {
	awsSession := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(os.Getenv("REGION")),
	}))

	dbs := dynamodb.New(awsSession)

	// Create a map of DynamoDB Attribute Values containing the table keys
	km := make(map[string]*dynamodb.AttributeValue)
	km[":productid"] = &dynamodb.AttributeValue{
		S: aws.String(productID),
	}

	si := &dynamodb.ScanInput{
		TableName:                 aws.String(os.Getenv("TABLE")),
		ExpressionAttributeValues: km,
		FilterExpression:          aws.String("ID = :productid"),
	}

	so, err := dbs.Scan(si)
	if err != nil {
		return catalog.Product{}, err
	}

	if len(so.Items) == 0 {
		return catalog.Product{}, fmt.Errorf("Unable to find product with ID %s", productID)
	}

	str := *so.Items[0]["ProductContent"].S
	prod, err := catalog.UnmarshalProduct(str)
	if err != nil {
		return catalog.Product{}, err
	}
	return prod, nil
}

func (m manager) GetProducts() ([]catalog.Product, error) {
	awsSession := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(os.Getenv("REGION")),
	}))

	dbs := dynamodb.New(awsSession)

	si := &dynamodb.ScanInput{
		TableName: aws.String(os.Getenv("TABLE")),
	}

	so, err := dbs.Scan(si)
	if err != nil {
		return nil, err
	}

	prods := make([]catalog.Product, len(so.Items))

	for idx, ct := range so.Items {
		str := *ct["ProductContent"].S
		prod, err := catalog.UnmarshalProduct(str)
		if err != nil {
			errormessage := fmt.Sprintf("error unmarshalling product data: %s", err.Error())
			log.Println(errormessage)
			break
		}
		prods[idx] = prod
	}

	return prods, nil
}
