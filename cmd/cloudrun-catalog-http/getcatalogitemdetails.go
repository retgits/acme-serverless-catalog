package main

import (
	"net/http"

	"github.com/valyala/fasthttp"
)

// GetCatalogItemDetails ...
func GetCatalogItemDetails(ctx *fasthttp.RequestCtx) {
	// Create the key attributes
	productID := ctx.UserValue("id").(string)

	// Get all products from the catalog
	prod, err := db.GetProduct(productID)
	if err != nil {
		ErrorHandler(ctx, "GetCatalogItemDetails", "GetProduct", err)
		return
	}

	payload, err := prod.Marshal()
	if err != nil {
		ErrorHandler(ctx, "GetCatalogItemDetails", "Marshal", err)
		return
	}

	ctx.SetStatusCode(http.StatusOK)
	ctx.Write(payload)
}
