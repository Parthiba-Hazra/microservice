package utils

import (
	"net/http"
	"time"
)

var ProductServiceURL string

func InitHTTPClient() {
	ProductServiceURL = getEnv("PRODUCT_SERVICE_URL", "http://localhost:8082")
}

var HTTPClient = &http.Client{
	Timeout: 10 * time.Second,
}
