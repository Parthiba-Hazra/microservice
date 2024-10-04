package graph

import (
	"context"
	"graphql-gateway/graph/model"
)

// Users is the resolver for the users field.
func (r *queryResolver) Users(ctx context.Context) ([]*model.User, error) {
	return r.Resolver.Users(ctx)
}

// User is the resolver for the user field.
func (r *queryResolver) User(ctx context.Context, id string) (*model.User, error) {
	return r.Resolver.User(ctx, id)
}

// Products is the resolver for the products field.
func (r *queryResolver) Products(ctx context.Context) ([]*model.Product, error) {
	return r.Resolver.Products(ctx)
}

// Product is the resolver for the product field.
func (r *queryResolver) Product(ctx context.Context, id string) (*model.Product, error) {
	return r.Resolver.Product(ctx, id)
}

// Orders is the resolver for the orders field.
func (r *queryResolver) Orders(ctx context.Context) ([]*model.Order, error) {
	return r.Resolver.Orders(ctx)
}

// Order is the resolver for the order field.
func (r *queryResolver) Order(ctx context.Context, id string) (*model.Order, error) {
	return r.Resolver.Order(ctx, id)
}

// RegisterUser is the resolver for the registerUser field.
func (r *mutationResolver) RegisterUser(ctx context.Context, input model.RegisterInput) (*model.RegisterUserResponse, error) {
	return r.Resolver.RegisterUser(ctx, input)
}

// CreateProduct is the resolver for the createProduct field.
func (r *mutationResolver) CreateProduct(ctx context.Context, input model.ProductInput) (*model.ProductResponse, error) {
	return r.Resolver.CreateProduct(ctx, input)
}

// PlaceOrder is the resolver for the placeOrder field.
func (r *mutationResolver) PlaceOrder(ctx context.Context, input model.OrderInput) (*model.OrderResponse, error) {
	return r.Resolver.PlaceOrder(ctx, input)
}

// Mutation returns MutationResolver implementation.
func (r *Resolver) Mutation() MutationResolver { return &mutationResolver{r} }

// Query returns QueryResolver implementation.
func (r *Resolver) Query() QueryResolver { return &queryResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
