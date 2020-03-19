package main

import (
	"fmt"
	"os"
	"path"

	"github.com/pulumi/pulumi-aws/sdk/go/aws/apigateway"
	"github.com/pulumi/pulumi-aws/sdk/go/aws/dynamodb"
	"github.com/pulumi/pulumi-aws/sdk/go/aws/iam"
	"github.com/pulumi/pulumi-aws/sdk/go/aws/lambda"
	"github.com/pulumi/pulumi/sdk/go/pulumi"
	"github.com/pulumi/pulumi/sdk/go/pulumi/config"
	"github.com/retgits/pulumi-helpers/builder"
	"github.com/retgits/pulumi-helpers/sampolicies"
)

// Tags are key-value pairs to apply to the resources created by this stack
type Tags struct {
	// Author is the person who created the code, or performed the deployment
	Author pulumi.String

	// Feature is the project that this resource belongs to
	Feature pulumi.String

	// Team is the team that is responsible to manage this resource
	Team pulumi.String

	// Version is the version of the code for this resource
	Version pulumi.String
}

// GenericConfig contains the key-value pairs for the configuration of AWS in this stack
type GenericConfig struct {
	// The AWS region used
	Region string

	// The DSN used to connect to Sentry
	SentryDSN string `json:"sentrydsn"`

	// The AWS AccountID to use
	AccountID string `json:"accountid"`
}

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		// Get the region
		region, found := ctx.GetConfig("aws:region")
		if !found {
			return fmt.Errorf("region not found")
		}

		// Read the configuration data from Pulumi.<stack>.yaml
		conf := config.New(ctx, "awsconfig")

		// Create a new Tags object with the data from the configuration
		var tags Tags
		conf.RequireObject("tags", &tags)

		// Create a new GenericConfig object with the data from the configuration
		var genericConfig GenericConfig
		conf.RequireObject("generic", &genericConfig)
		genericConfig.Region = region

		// Create a map[string]pulumi.Input of the tags
		// the first four tags come from the configuration file
		// the last two are derived from this deployment
		tagMap := make(map[string]pulumi.Input)
		tagMap["Author"] = tags.Author
		tagMap["Feature"] = tags.Feature
		tagMap["Team"] = tags.Team
		tagMap["Version"] = tags.Version
		tagMap["ManagedBy"] = pulumi.String("Pulumi")
		tagMap["Stage"] = pulumi.String(ctx.Stack())

		// functions are the functions that need to be deployed
		functions := []string{
			"lambda-catalog-all",
			"lambda-catalog-get",
			"lambda-catalog-newproduct",
		}

		// Compile and zip the AWS Lambda functions
		wd, err := os.Getwd()
		if err != nil {
			return err
		}

		for _, fnName := range functions {
			// Find the working folder
			fnFolder := path.Join(wd, "..", "cmd", fnName)
			buildFactory := builder.NewFactory().WithFolder(fnFolder)
			buildFactory.MustBuild()
			buildFactory.MustZip()
		}

		// Create a factory to get policies from
		iamFactory := sampolicies.NewFactory().WithAccountID(genericConfig.AccountID).WithPartition("aws").WithRegion(genericConfig.Region)

		// Create the API Gateway Policy
		iamFactory.AddAssumeRoleLambda()
		iamFactory.AddExecuteAPI()
		policies, err := iamFactory.GetPolicyStatement()
		if err != nil {
			return err
		}

		// Create an API Gateway
		gateway, err := apigateway.NewRestApi(ctx, "CatalogService", &apigateway.RestApiArgs{
			Name:        pulumi.String("CatalogService"),
			Description: pulumi.String("ACME Serverless Fitness Shop - Catalog"),
			Tags:        pulumi.Map(tagMap),
			Policy:      pulumi.String(policies),
		})
		if err != nil {
			return err
		}

		// Create the parent resources in the API Gateway
		productsResource, err := apigateway.NewResource(ctx, "ProductsAPIResource", &apigateway.ResourceArgs{
			RestApi:  gateway.ID(),
			PathPart: pulumi.String("products"),
			ParentId: gateway.RootResourceId,
		})
		if err != nil {
			return err
		}

		// Lookup the DynamoDB table
		dynamoTable, err := dynamodb.LookupTable(ctx, &dynamodb.LookupTableArgs{
			Name: fmt.Sprintf("%s-acmeserverless-dynamodb", ctx.Stack()),
		})

		// dynamoPolicy is a policy template, derived from AWS SAM, to allow apps
		// to connect to and execute command on Amazon DynamoDB
		iamFactory.ClearPolicies()
		iamFactory.AddDynamoDBCrudPolicy(dynamoTable.Name)
		dynamoPolicy, err := iamFactory.GetPolicyStatement()
		if err != nil {
			return err
		}

		roles := make(map[string]*iam.Role)

		// Create a new IAM role for each Lambda function
		for _, function := range functions {
			// Give the role the ability to run on AWS Lambda
			roleArgs := &iam.RoleArgs{
				AssumeRolePolicy: pulumi.String(sampolicies.AssumeRoleLambda()),
				Description:      pulumi.String(fmt.Sprintf("Role for the Catalog Service (%s) of the ACME Serverless Fitness Shop", function)),
				Tags:             pulumi.Map(tagMap),
			}

			role, err := iam.NewRole(ctx, fmt.Sprintf("ACMEServerlessCatalogRole-%s", function), roleArgs)
			if err != nil {
				return err
			}

			// Attach the AWSLambdaBasicExecutionRole so the function can create Log groups in CloudWatch
			_, err = iam.NewRolePolicyAttachment(ctx, fmt.Sprintf("AWSLambdaBasicExecutionRole-%s", function), &iam.RolePolicyAttachmentArgs{
				PolicyArn: pulumi.String("arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"),
				Role:      role.Name,
			})
			if err != nil {
				return err
			}

			// Add the DynamoDB policy
			_, err = iam.NewRolePolicy(ctx, fmt.Sprintf("ACMEServerlessCatalogPolicy-%s", function), &iam.RolePolicyArgs{
				Name:   pulumi.String(fmt.Sprintf("ACMEServerlessCatalogPolicy-%s", function)),
				Role:   role.Name,
				Policy: pulumi.String(dynamoPolicy),
			})
			if err != nil {
				return err
			}

			ctx.Export(fmt.Sprintf("%s-role::Arn", function), role.Arn)
			roles[function] = role
		}

		// All functions will have the same environment variables, with the exception
		// of the function name
		variables := make(map[string]pulumi.StringInput)
		variables["REGION"] = pulumi.String(genericConfig.Region)
		variables["SENTRY_DSN"] = pulumi.String(genericConfig.SentryDSN)
		variables["VERSION"] = tags.Version
		variables["STAGE"] = pulumi.String(ctx.Stack())
		variables["TABLE"] = pulumi.String(dynamoTable.Name)

		variables["FUNCTION_NAME"] = pulumi.String(fmt.Sprintf("%s-lambda-catalog-all", ctx.Stack()))
		environment := lambda.FunctionEnvironmentArgs{
			Variables: pulumi.StringMap(variables),
		}

		// Create the All function
		functionArgs := &lambda.FunctionArgs{
			Description: pulumi.String("A Lambda function to get all products from DynamoDB"),
			Runtime:     pulumi.String("go1.x"),
			Name:        pulumi.String(fmt.Sprintf("%s-lambda-catalog-all", ctx.Stack())),
			MemorySize:  pulumi.Int(256),
			Timeout:     pulumi.Int(10),
			Handler:     pulumi.String("lambda-catalog-all"),
			Environment: environment,
			Code:        pulumi.NewFileArchive("../cmd/lambda-catalog-all/lambda-catalog-all.zip"),
			Role:        roles["lambda-catalog-all"].Arn,
			Tags:        pulumi.Map(tagMap),
		}

		function, err := lambda.NewFunction(ctx, fmt.Sprintf("%s-lambda-catalog-all", ctx.Stack()), functionArgs)
		if err != nil {
			return err
		}

		_, err = apigateway.NewMethod(ctx, "AllCatalogsAPIGetMethod", &apigateway.MethodArgs{
			HttpMethod:    pulumi.String("GET"),
			Authorization: pulumi.String("NONE"),
			RestApi:       gateway.ID(),
			ResourceId:    productsResource.ID(),
		}, pulumi.DependsOn([]pulumi.Resource{gateway, productsResource}))
		if err != nil {
			return err
		}

		_, err = apigateway.NewIntegration(ctx, "AllCatalogsAPIIntegration", &apigateway.IntegrationArgs{
			HttpMethod:            pulumi.String("GET"),
			IntegrationHttpMethod: pulumi.String("POST"),
			ResourceId:            productsResource.ID(),
			RestApi:               gateway.ID(),
			Type:                  pulumi.String("AWS_PROXY"),
			Uri:                   function.InvokeArn,
		}, pulumi.DependsOn([]pulumi.Resource{gateway, productsResource, function}))
		if err != nil {
			return err
		}

		_, err = lambda.NewPermission(ctx, "AllCatalogsAPIPermission", &lambda.PermissionArgs{
			Action:    pulumi.String("lambda:InvokeFunction"),
			Function:  function.Name,
			Principal: pulumi.String("apigateway.amazonaws.com"),
			SourceArn: pulumi.Sprintf("arn:aws:execute-api:%s:%s:%s/*/GET/products", genericConfig.Region, genericConfig.AccountID, gateway.ID()),
		}, pulumi.DependsOn([]pulumi.Resource{gateway, productsResource, function}))
		if err != nil {
			return err
		}

		ctx.Export("lambda-catalog-all::Arn", function.Arn)

		// Create the Get function
		variables["FUNCTION_NAME"] = pulumi.String(fmt.Sprintf("%s-lambda-catalog-get", ctx.Stack()))
		environment = lambda.FunctionEnvironmentArgs{
			Variables: pulumi.StringMap(variables),
		}

		functionArgs = &lambda.FunctionArgs{
			Description: pulumi.String("A Lambda function to get a single product from DynamoDB"),
			Runtime:     pulumi.String("go1.x"),
			Name:        pulumi.String(fmt.Sprintf("%s-lambda-catalog-get", ctx.Stack())),
			MemorySize:  pulumi.Int(256),
			Timeout:     pulumi.Int(10),
			Handler:     pulumi.String("lambda-catalog-get"),
			Environment: environment,
			Code:        pulumi.NewFileArchive("../cmd/lambda-catalog-get/lambda-catalog-get.zip"),
			Role:        roles["lambda-catalog-get"].Arn,
			Tags:        pulumi.Map(tagMap),
		}

		function, err = lambda.NewFunction(ctx, fmt.Sprintf("%s-lambda-catalog-get", ctx.Stack()), functionArgs)
		if err != nil {
			return err
		}

		resource, err := apigateway.NewResource(ctx, "GetCatalogsAPI", &apigateway.ResourceArgs{
			RestApi:  gateway.ID(),
			PathPart: pulumi.String("{id}"),
			ParentId: productsResource.ID(),
		}, pulumi.DependsOn([]pulumi.Resource{gateway}))
		if err != nil {
			return err
		}

		_, err = apigateway.NewMethod(ctx, "GetCatalogsAPIGetMethod", &apigateway.MethodArgs{
			HttpMethod:    pulumi.String("GET"),
			Authorization: pulumi.String("NONE"),
			RestApi:       gateway.ID(),
			ResourceId:    resource.ID(),
		}, pulumi.DependsOn([]pulumi.Resource{gateway, resource}))
		if err != nil {
			return err
		}

		_, err = apigateway.NewIntegration(ctx, "GetCatalogsAPIIntegration", &apigateway.IntegrationArgs{
			HttpMethod:            pulumi.String("GET"),
			IntegrationHttpMethod: pulumi.String("POST"),
			ResourceId:            resource.ID(),
			RestApi:               gateway.ID(),
			Type:                  pulumi.String("AWS_PROXY"),
			Uri:                   function.InvokeArn,
		}, pulumi.DependsOn([]pulumi.Resource{gateway, resource, function}))
		if err != nil {
			return err
		}

		_, err = lambda.NewPermission(ctx, "GetCatalogsAPIPermission", &lambda.PermissionArgs{
			Action:    pulumi.String("lambda:InvokeFunction"),
			Function:  function.Name,
			Principal: pulumi.String("apigateway.amazonaws.com"),
			SourceArn: pulumi.Sprintf("arn:aws:execute-api:%s:%s:%s/*/GET/products/*", genericConfig.Region, genericConfig.AccountID, gateway.ID()),
		}, pulumi.DependsOn([]pulumi.Resource{gateway, resource, function}))
		if err != nil {
			return err
		}

		ctx.Export("lambda-catalog-get::Arn", function.Arn)

		// Create the NewProduct function
		variables["FUNCTION_NAME"] = pulumi.String(fmt.Sprintf("%s-lambda-catalog-newproduct", ctx.Stack()))
		environment = lambda.FunctionEnvironmentArgs{
			Variables: pulumi.StringMap(variables),
		}

		functionArgs = &lambda.FunctionArgs{
			Description: pulumi.String("A Lambda function to create new products in DynamoDB"),
			Runtime:     pulumi.String("go1.x"),
			Name:        pulumi.String(fmt.Sprintf("%s-lambda-catalog-newproduct", ctx.Stack())),
			MemorySize:  pulumi.Int(256),
			Timeout:     pulumi.Int(10),
			Handler:     pulumi.String("lambda-catalog-newproduct"),
			Environment: environment,
			Code:        pulumi.NewFileArchive("../cmd/lambda-catalog-newproduct/lambda-catalog-newproduct.zip"),
			Role:        roles["lambda-catalog-newproduct"].Arn,
			Tags:        pulumi.Map(tagMap),
		}
		variables["FUNCTION_NAME"] = pulumi.String(fmt.Sprintf("%s-lambda-catalog-newproduct", ctx.Stack()))

		function, err = lambda.NewFunction(ctx, fmt.Sprintf("%s-lambda-catalog-newproduct", ctx.Stack()), functionArgs)
		if err != nil {
			return err
		}

		resource, err = apigateway.NewResource(ctx, "NewCatalogAPI", &apigateway.ResourceArgs{
			RestApi:  gateway.ID(),
			PathPart: pulumi.String("product"),
			ParentId: gateway.RootResourceId,
		}, pulumi.DependsOn([]pulumi.Resource{gateway}))
		if err != nil {
			return err
		}

		_, err = apigateway.NewMethod(ctx, "NewCatalogAPIPostMethod", &apigateway.MethodArgs{
			HttpMethod:    pulumi.String("POST"),
			Authorization: pulumi.String("NONE"),
			RestApi:       gateway.ID(),
			ResourceId:    resource.ID(),
		}, pulumi.DependsOn([]pulumi.Resource{gateway, resource}))
		if err != nil {
			return err
		}

		_, err = apigateway.NewIntegration(ctx, "NewCatalogAPIIntegration", &apigateway.IntegrationArgs{
			HttpMethod:            pulumi.String("POST"),
			IntegrationHttpMethod: pulumi.String("POST"),
			ResourceId:            resource.ID(),
			RestApi:               gateway.ID(),
			Type:                  pulumi.String("AWS_PROXY"),
			Uri:                   function.InvokeArn,
		}, pulumi.DependsOn([]pulumi.Resource{gateway, resource, function}))
		if err != nil {
			return err
		}

		_, err = lambda.NewPermission(ctx, "NewCatalogAPIPermission", &lambda.PermissionArgs{
			Action:    pulumi.String("lambda:InvokeFunction"),
			Function:  function.Name,
			Principal: pulumi.String("apigateway.amazonaws.com"),
			SourceArn: pulumi.Sprintf("arn:aws:execute-api:%s:%s:%s/*/POST/product", genericConfig.Region, genericConfig.AccountID, gateway.ID()),
		}, pulumi.DependsOn([]pulumi.Resource{gateway, resource, function}))
		if err != nil {
			return err
		}

		ctx.Export("lambda-catalog-newproduct::Arn", function.Arn)

		return nil
	})
}
