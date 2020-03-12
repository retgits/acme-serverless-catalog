# Catalog

> A catalog service, because what is a shop without a catalog to show off our awesome red pants?

The Catalog service is part of the [ACME Fitness Serverless Shop](https://github.com/retgits/acme-serverless). The goal of this specific service is to register and serve the catalog of items sold by the shop.

## Prerequisites

* [Go (at least Go 1.12)](https://golang.org/dl/)
* [An AWS account](https://portal.aws.amazon.com/billing/signup)
* [A Pulumi account](https://app.pulumi.com/signup)
* [A Sentry.io account](https://sentry.io) if you want to enable tracing and error reporting

## Deploying

To deploy the Catalog Service you'll need a [Pulumi account](https://app.pulumi.com/signup). Once you have your Pulumi account and configured the [Pulumi CLI](https://www.pulumi.com/docs/get-started/aws/install-pulumi/), you can initialize a new stack using the Pulumi templates in the [pulumi](./pulumi) folder.

```bash
cd pulumi
pulumi stack init <your pulumi org>/acmeserverless-catalog/dev
```

You'll need to create a [Pulumi.dev.yaml](./pulumi/Pulumi.dev.yaml) file that will contain all configuration data to deploy the app:

```yaml
config:
  aws:region: us-west-2 ## The region you want to deploy to
  awsconfig:lambda:
    dynamoarn: ## The ARN to the DynamoDB table
    sentrydsn: ## The DSN to connect to Sentry
    region: ## The region you want to deploy to
    accountid: ## Your AWS sccount ID
  awsconfig:tags:
    author: retgits ## The author, you...
    feature: acmeserverless
    team: vcs ## The team you're on
    version: 0.1.0 ## The version
```

To create the Pulumi stack, and create the Catalog service, run `pulumi up`.

If you want to keep track of the resources in Pulumi, you can add tags to your stack as well.

```bash
pulumi stack tag set app:name acmeserverless
pulumi stack tag set app:feature acmeserverless-catalog
pulumi stack tag set app:domain catalog
```

## API

### `GET /products`

Returns a list of all catalog items

```bash
curl --request GET \
  --url https://<id>.execute-api.us-west-2.amazonaws.com/Prod/products
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
  --url https://<id>.execute-api.us-west-2.amazonaws.com/Prod/products \
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
  --url https://<id>.execute-api.us-west-2.amazonaws.com/Prod/products/5c61f497e5fdadefe84ff9b9
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
