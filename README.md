# Catalog

> A catalog service, because what is a shop without a catalog to show off our awesome red pants?

The Catalog service is part of the [ACME Fitness Serverless Shop](https://github.com/vmwarecloudadvocacy/acme_fitness_demo). The goal of this specific service is to register and serve the catalog of items sold by the shop.

## Prerequisites

* [Go (at least Go 1.12)](https://golang.org/dl/)
* [An AWS Account](https://portal.aws.amazon.com/billing/signup)
* The _vuln_ targets for Make and Mage rely on the [Snyk](http://snyk.io/) CLI

## Eventing Options

The catalog service has Lambdas triggered by [Amazon API Gateway](https://aws.amazon.com/api-gateway/)

## Data Stores

The order service supports the following data stores:

* [Amazon DynamoDB](https://aws.amazon.com/dynamodb/). The table can be created using the makefile in [create-dynamodb](./cmd/create-dynamodb).

## Using Amazon API Gateway

### Prerequisites for Amazon API Gateway

* [AWS CLI](https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-install.html) installed and configured

### Build and deploy for Amazon API Gateway

Clone this repository

```bash
git clone https://github.com/retgits/acme-serverless-catalog
cd acme-serverless-catalog
```

Get the Go Module dependencies

```bash
go get ./...
```

Switch directories to any of the Lambda folders

```bash
cd ./cmd/lambda-catalog-<name>
```

Use make to deploy

```bash
make build
make deploy
```

### Testing Amazon API Gateway

After the deployment you'll see the URL to which you can send the below mentioned API requests

## API

### `GET /products`

Returns a list of all catalog items

```bash
curl --request GET \
  --url http://localhost:8082/products
```

```json
{
    "data": [
    {
      "id": "5c61f497e5fdadefe84ff9b9",
      "name": "Yoga Mat",
      "shortDescription": "Limited Edition Mat",
      "description": "Limited edition yoga mat",
      "imageUrl1": "/static/images/yogamat_square.jpg",
      "imageUrl2": "/static/images/yogamat_thumb2.jpg",
      "imageUrl3": "/static/images/yogamat_thumb3.jpg",
      "price": 62.5,
      "tags": [
          "mat"
      ]
    },
    {
      "id": "5c61f497e5fdadefe84ff9ba",
      "name": "Water Bottle",
      "shortDescription": "Best water bottle ever",
      "description": "For all those athletes out there, a perfect bottle to enrich you",
      "imageUrl1": "/static/images/bottle_square.jpg",
      "imageUrl2": "/static/images/bottle_thumb2.jpg",
      "imageUrl3": "/static/images/bottle_thumb3.jpg",
      "price": 34.99,
      "tags": [
          "bottle"
          ]
    }
    ]}
```

### `POST /product`

Create a new product item

```bash
curl --request POST \
  --url http://localhost:8082/products \
  --header 'content-type: application/json' \
  --data '         {
            "name": "Tracker",
            "shortDescription": "Limited Edition Tracker",
            "description": "Limited edition Tracker with longer description",
            "imageurl1": "/static/images/tracker_square.jpg",
            "imageurl2": "/static/images/tracker_thumb2.jpg",
            "imageurl3": "/static/images/tracker_thumb3.jpg",
            "price": 149.99,
            "tags": [
                "tracker"
             ]

          }'
```

The call to this service needs a valid product object

```json
{
    "name": "Tracker",
    "shortDescription": "Limited Edition Tracker",
    "description": "Limited edition Tracker with longer description",
    "imageurl1": "/static/images/tracker_square.jpg",
    "imageurl2": "/static/images/tracker_thumb2.jpg",
    "imageurl3": "/static/images/tracker_thumb3.jpg",
    "price": 149.99,
    "tags": [
        "tracker"
    ]
}
```

When the product is created successfully, an HTTP/201 message is returned

```json
{
    "message": "Product created successfully!",
    "resourceId": {
        "id": "5c61f8f81d41c8e94ecaf25f",
        "name": "Tracker",
        "shortDescription": "Limited Edition Tracker",
        "description": "Limited edition Tracker with longer description",
        "imageUrl1": "/static/images/tracker_square.jpg",
        "imageUrl2": "/static/images/tracker_thumb2.jpg",
        "imageUrl3": "/static/images/tracker_thumb3.jpg",
        "price": 149.99,
        "tags": [
            "tracker"
        ]
    },
    "status": 201
}
```

### `GET /products/:id`

Returns details about a specific product id

```bash
curl --request GET \
  --url http://localhost:8082/products/5c61f497e5fdadefe84ff9b9
```

```json
{
    "data": {
        "id": "5c61f497e5fdadefe84ff9b9",
        "name": "Yoga Mat",
        "shortDescription": "Limited Edition Mat",
        "description": "Limited edition yoga mat",
        "imageUrl1": "/static/images/yogamat_square.jpg",
        "imageUrl2": "/static/images/yogamat_square.jpg",
        "imageUrl3": "/static/images/bottle_square.jpg",
        "price": 62.5,
        "tags": [
            "mat"
        ]
    },
    "status": 200
}
```

## Contributing

[Pull requests](https://github.com/retgits/acme-serverless-catalog/pulls) are welcome. For major changes, please open [an issue](https://github.com/retgits/acme-serverless-catalog/issues) first to discuss what you would like to change.

Please make sure to update tests as appropriate.

## License

See the [LICENSE](./LICENSE) file in the repository