// Package mongodb leverages cross-platform document-oriented database program. Classified as a
// NoSQL database program, MongoDB uses JSON-like documents with schema.
package mongodb

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	acmeserverless "github.com/retgits/acme-serverless"
	"github.com/retgits/acme-serverless-catalog/internal/datastore"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// The pointer to MongoDB provides the API operation methods for making requests to MongoDB.
// This specifically creates a single instance of the MongoDB service which can be reused if the
// container stays warm.
var dbs *mongo.Collection

// manager is an empty struct that implements the methods of the
// Manager interface.
type manager struct{}

// init creates the connection to MongoDB.
func init() {
	username := os.Getenv("MONGO_USERNAME")
	password := os.Getenv("MONGO_PASSWORD")
	hostname := os.Getenv("MONGO_HOSTNAME")
	port := os.Getenv("MONGO_PORT")

	connString := fmt.Sprintf("mongodb+srv://%s:%s@%s:%s", username, password, hostname, port)
	if strings.HasSuffix(connString, ":") {
		connString = connString[:len(connString)-1]
	}
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(connString))
	if err != nil {
		log.Fatalf("error connecting to MongoDB: %s", err.Error())
	}
	dbs = client.Database("acmeserverless").Collection("catalog")
}

// New creates a new datastore manager using Amazon DynamoDB as backend
func New() datastore.Manager {
	return manager{}
}

// AddProduct stores a new product in Amazon DynamoDB
func (m manager) AddProduct(p acmeserverless.CatalogItem) error {
	payload, err := p.Marshal()
	if err != nil {
		return err
	}

	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	_, err = dbs.InsertOne(ctx, bson.D{{"SK", p.ID}, {"PK", "PRODUCT"}, {"Payload", string(payload)}})

	return err
}

// GetProduct retrieves a single product from DynamoDB based on the productID
func (m manager) GetProduct(productID string) (acmeserverless.CatalogItem, error) {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)

	res := dbs.FindOne(ctx, bson.D{{"SK", productID}})

	raw, err := res.DecodeBytes()
	if err != nil {
		return acmeserverless.CatalogItem{}, fmt.Errorf("unable to decode bytes: %s", err.Error())
	}

	return acmeserverless.UnmarshalCatalogItem(raw.Lookup("Payload").StringValue())
}

// GetProducts retrieves all products from DynamoDB
func (m manager) GetProducts() ([]acmeserverless.CatalogItem, error) {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	cursor, err := dbs.Find(ctx, bson.D{})
	if err != nil {
		log.Fatal(err)
	}

	var results []bson.M

	if err = cursor.All(ctx, &results); err != nil {
		log.Fatal(err)
	}

	prods := make([]acmeserverless.CatalogItem, 0)

	for _, result := range results {
		prod, err := acmeserverless.UnmarshalCatalogItem(result["Payload"].(string))
		if err != nil {
			log.Println(fmt.Sprintf("error unmarshalling catalog item data: %s", err.Error()))
			continue
		}

		prods = append(prods, prod)
	}

	return prods, nil
}
