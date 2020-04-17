package main

import (
	"net/http"

	acmeserverless "github.com/retgits/acme-serverless"
	"github.com/valyala/fasthttp"
)

// GetAllCatalogItems ...
func GetAllCatalogItems(ctx *fasthttp.RequestCtx) {
	// Get all products from the catalog
	products, err := db.GetProducts()
	if err != nil {
		ErrorHandler(ctx, "GetAllCatalogItems", "GetProducts", err)
		return
	}

	res := acmeserverless.AllCatalogItemsResponse{
		Data: products,
	}

	payload, err := res.Marshal()
	if err != nil {
		ErrorHandler(ctx, "GetAllCatalogItems", "Marshal", err)
		return
	}

	ctx.SetStatusCode(http.StatusOK)
	ctx.Write(payload)
}
