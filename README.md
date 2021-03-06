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

Pulumi is configured using a file called `Pulumi.dev.yaml`. A sample configuration is available in the Pulumi directory. You can rename [`Pulumi.dev.yaml.sample`](./pulumi/Pulumi.dev.yaml.sample) to `Pulumi.dev.yaml` and update the variables accordingly. Alternatively, you can change variables directly in the [main.go](./pulumi/main.go) file in the pulumi directory. The configuration contains:

```yaml
config:
  aws:region: us-west-2 ## The region you want to deploy to
  awsconfig:generic:
    sentrydsn: ## The DSN to connect to Sentry
    accountid: ## Your AWS Account ID
    wavefronturl: ## The URL of your Wavefront instance
    wavefronttoken: ## Your Wavefront API token
  awsconfig:tags:
    author: retgits ## The author, you...
    feature: acmeserverless
    team: vcs ## The team you're on
    version: 0.2.0 ## The version
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

## Building for Google Cloud Run

If you have Docker installed locally, you can use `docker build` to create a container which can be used to try out the catalog service locally and for Google Cloud Run.

To build your container image using Docker:

Run the command:

```bash
VERSION=`git describe --tags --always --dirty="-dev"`
docker build -f ./cmd/cloudrun-catalog-http/Dockerfile . -t gcr.io/[PROJECT-ID]/catalog:$VERSION
```

Replace `[PROJECT-ID]` with your Google Cloud project ID

If you have not yet configured Docker to use the gcloud command-line tool to authenticate requests to Container Registry, do so now using the command:

```bash
gcloud auth configure-docker
```

You need to do this before you can push or pull images using Docker. You only need to do it once.

Push the container image to Container Registry:

```bash
docker push gcr.io/[PROJECT-ID]/catalog:$VERSION
```

The container relies on the environment variables:

* SENTRY_DSN: The DSN to connect to Sentry
* K_SERVICE: The name of the service (in Google Cloud Run this variable is automatically set, defaults to `catalog` if not set)
* VERSION: The version you're running (will default to `dev` if not set)
* PORT: The port number the service will listen on (will default to `8080` if not set)
* STAGE: The environment in which you're running
* WAVEFRONT_TOKEN: The token to connect to Wavefront
* WAVEFRONT_URL: The URL to connect to Wavefront (will default to `debug` if not set)
* MONGO_USERNAME: The username to connect to MongoDB
* MONGO_PASSWORD: The password to connect to MongoDB
* MONGO_HOSTNAME: The hostname of the MongoDB server
* MONGO_PORT: The port number of the MongoDB server

A `docker run`, with all options, is:

```bash
docker run --rm -it -p 8080:8080 -e SENTRY_DSN=abcd -e K_SERVICE=catalog \
  -e VERSION=$VERSION -e PORT=8080 -e STAGE=dev -e WAVEFRONT_URL=https://my-url.wavefront.com \
  -e WAVEFRONT_TOKEN=efgh -e MONGO_USERNAME=admin -e MONGO_PASSWORD=admin \
  -e MONGO_HOSTNAME=localhost -e MONGO_PORT=27017 gcr.io/[PROJECT-ID]/catalog:$VERSION
```

Replace `[PROJECT-ID]` with your Google Cloud project ID

## Troubleshooting

In case the API Gateway responds with `{"message":"Forbidden"}`, there is likely an issue with the deployment of the API Gateway. To solve this problem, you can use the AWS CLI. To confirm this, run `aws apigateway get-deployments --rest-api-id <rest-api-id>`. If that returns no deployments, you can create a deployment for the *prod* stage with `aws apigateway create-deployment --rest-api-id <rest-api-id> --stage-name prod --stage-description 'Prod Stage' --description 'deployment to the prod stage'`.

## Contributing

[Pull requests](https://github.com/retgits/acme-serverless-catalog/pulls) are welcome. For major changes, please open [an issue](https://github.com/retgits/acme-serverless-catalog/issues) first to discuss what you would like to change.

Please make sure to update tests as appropriate.

## License

See the [LICENSE](./LICENSE) file in the repository
