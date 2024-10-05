package main

import (
	"context"
	"graphql-gateway/cache"
	"graphql-gateway/graph"
	"net/http"
	"os"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/go-chi/chi"
)

const defaultPort = "8080"

func main() {
	// Initialize Redis client
	cache.InitRedis()

	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	router := chi.NewRouter()

	// Middleware to add Authorization header to context
	router.Use(authorizationMiddleware)

	// GraphQL handler
	srv := handler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{Resolvers: &graph.Resolver{}}))

	router.Handle("/", playground.Handler("GraphQL playground", "/query"))
	router.Handle("/query", srv)

	http.ListenAndServe(":"+port, router)
}

func authorizationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract the Authorization header from the request
		authHeader := r.Header.Get("Authorization")

		// Add it to the context
		ctx := context.WithValue(r.Context(), "Authorization", authHeader)

		// Pass the updated context to the next handler
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
