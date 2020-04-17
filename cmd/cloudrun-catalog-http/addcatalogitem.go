package main

import (
	"net/http"

	"github.com/gofrs/uuid"
	acmeserverless "github.com/retgits/acme-serverless"
	"github.com/valyala/fasthttp"
)

// AddCatalogItem ...
func AddCatalogItem(ctx *fasthttp.RequestCtx) {
	// Update the product with an ID
	prod, err := acmeserverless.UnmarshalCatalogItem(string(ctx.Request.Body()))
	if err != nil {
		ErrorHandler(ctx, "AddCatalogItem", "UnmarshalCatalogItem", err)
		return
	}
	prod.ID = uuid.Must(uuid.NewV4()).String()

	// Store a new product in the catalog
	err = db.AddProduct(prod)
	if err != nil {
		ErrorHandler(ctx, "AddCatalogItem", "AddProduct", err)
		return
	}

	status := acmeserverless.CreateCatalogItemResponse{
		Message:    "Product created successfully!",
		ResourceID: prod,
		Status:     http.StatusOK,
	}

	payload, err := status.Marshal()
	if err != nil {
		ErrorHandler(ctx, "AddCatalogItem", "Marshal", err)
		return
	}

	ctx.SetStatusCode(http.StatusOK)
	ctx.Write(payload)
}
