package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"

	catalog "github.com/retgits/acme-serverless-catalog"
	"github.com/retgits/acme-serverless-catalog/internal/datastore/dynamodb"
)

func main() {
	os.Setenv("REGION", "us-west-2")
	os.Setenv("TABLE", "Catalog")

	data, err := ioutil.ReadFile("./data.json")
	if err != nil {
		log.Println(err)
	}

	var products []catalog.Product

	err = json.Unmarshal(data, &products)
	if err != nil {
		log.Println(err)
	}

	dynamoStore := dynamodb.New()

	for _, product := range products {
		err = dynamoStore.AddProduct(product)
		if err != nil {
			log.Println(err)
		}
	}
}
