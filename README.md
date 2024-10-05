
# Microservice Application

## Overview

This project contains three main micro-services and a GraphQL gateway:

1.  **User Service**: Manages user registration, login, and user details.
2.  **Product Service**: Manages products, including creation, updates, deletions, and product information.
3.  **Order Service**: Handles order creation, retrieval, and management.
4.  **GraphQL Gateway**: Unifies all services into a single endpoint for querying and managing data using GraphQL.

### User Service

-   **Endpoints**:
    -   **POST /register**: Register a new user.
    -   **POST /login**: Login a user and receive a JWT token.
    -   **GET /users**: Retrieve all users.
    -   **GET /users/{:user_id}**
        
        : Retrieve a user by ID.
    -   **PUT /profile**: Update user profile (authenticated users only).
-   **Prometheus Metrics Endpoint**:
    -   **GET /metrics**: Metrics for monitoring using Prometheus.

### Product Service

-   **Endpoints**:
    -   **GET /products**: Retrieve a list of all products.
    -   **GET /products/**
        
        : Retrieve product details by ID.
    -   **POST /products**: Create a new product (admin only).
    -   **PUT /products/**
        
        : Update a product (admin only).
    -   **DELETE /products/**
        
        : Delete a product (admin only).
-   **Prometheus Metrics Endpoint**:
    -   **GET /metrics**: Metrics for monitoring using Prometheus.

### Order Service

-   **Endpoints**:
    -   **POST /orders**: Place a new order.
    -   **GET /orders**: Retrieve orders for the authenticated user.
    -   **GET /orders/{:order_id}**
        
        : Retrieve specific order details by ID.
-   **Prometheus Metrics Endpoint**:
    -   **GET /metrics**: Metrics for monitoring using Prometheus.

### GraphQL Gateway

The GraphQL Gateway is the main access point for interacting with the entire system. It allows for operations on users, products, and orders via a GraphQL API.

-   **GraphQL Queries and Mutations**:
    -   **Queries**:
        -   `users`, `user(id: ID!)`
        -   `products`, `product(id: ID!)`
        -   `orders`, `order(id: ID!)`
    -   **Mutations**:
        -   `registerUser(input: RegisterInput!)`
        -   `createProduct(input: ProductInput!)` (Admin only)
        -   `placeOrder(input: OrderInput!)` (Authenticated users)

## Prerequisites

-   Docker and Docker Compose installed on your machine.
-   Git to clone the repository.

## Setup Instructions

1.  Clone the repository:
    
    
    `git clone https://github.com/Parthiba-Hazra/microservice.git`
    
    `cd microservice` 
    
2.  Build and run the services using Docker Compose:
    

    
    `docker-compose up --build` 
    
3.  Once all services are running, you can access them using the following ports:
    
    -   **User Service**: `http://localhost:8081`
    -   **Product Service**: `http://localhost:8082`
    -   **Order Service**: `http://localhost:8083`
    -   **GraphQL Gateway**: `http://localhost:8080`

## Prometheus Metrics

Each service has a Prometheus endpoint available at `/metrics`. You can scrape the metrics using Prometheus or visualize them using Grafana.

## Authentication

-   ***To perform admin-level operations (such as creating or updating products), you need to create a user with the username "admin."***
-   Use the `/register` and `/login` endpoints to obtain JWT tokens for authentication.
-   Provide the JWT token in the `Authorization` header in the format `Bearer YOUR_TOKEN` for protected endpoints.

## GraphQL Playground

The GraphQL Playground is available at `http://localhost:8080`. Below are the queries and mutations supported in the application.

### GraphQL Playground Queries & Mutations

#### Headers Configuration for Authentication

To authenticate requests, set the Authorization header:

`{
  "Authorization": "Bearer YOUR_TOKEN"
}` 

Replace `YOUR_TOKEN` with the correct JWT token obtained from the `/login` endpoint.

### User Queries

1.  **Retrieve All Users**
    
    
    ```
    query {
      users {
        id
        username
        email
        created_at
      }
    }
    ``` 
    
2.  **Get User By ID**
    
    
    ```
    query {
      user(id: "1") {
        id
        username
        email
        created_at
      }
    }
    ``` 
    

### Product Queries

1.  **Retrieve All Products**
    
    ```
    query {
      products {
        id
        name
        description
        price
        inventory
        created_at
      }
    }
    ``` 
    
2.  **Get Product By ID**
    
    ```
    query {
      product(id: "1") {
        id
        name
        description
        price
        inventory
        created_at
      }
    }
    ``` 
    

### Order Queries 
Retrieve order information associated with Authorization header

1.  **Retrieve All Orders**
    
    ```
    query {
      orders {
        id
        user_id
        status
        total
        created_at
        items {
          id
          order_id
          product_id
          quantity
          price
        }
      }
    }
    ``` 
    Header:
    `
    {
  "Authorization": "Bearer YOUR_TOKEN"
}
    `
    
2.  **Get Order By ID**
    
    ```
    query {
      order(id: "1") {
        id
        user_id
        status
        total
        created_at
        items {
          id
          order_id
          product_id
          quantity
          price
        }
      }
    }
    ``` 
    Header:
    `
    {
  "Authorization": "Bearer YOUR_TOKEN"
}
    `
    

### User Mutations

1.  **Register User**
    
    ```
    mutation {
      registerUser(input: { username: "admin", email: "admin@example.com", password: "adminpass" }) {
        message
      }
    }
    ``` 
    
2.  **Login User** (This mutation is performed via a REST endpoint, not GraphQL. Use `/login` to obtain the JWT token.)
    

### Product Mutations

1.  **Create Product** (Admin Only)
    
    ```
    mutation {
      createProduct(input: {
        name: "Product A",
        description: "This is a sample product",
        price: 19.99,
        inventory: 50
      }) {
        message
      }
    }
    ``` 
    Header:
    `
    {
  "Authorization": "Bearer ADMIN_TOKEN"
}
    `
    
2.  **Update Product** (Admin Only / via a REST endpoint)
    
    
    ```
    curl -X PUT http://localhost:8082/products/{:product_id} \
    -H 'Content-Type: application/json' \
    -H 'Authorization: Bearer YOUR_ADMIN_TOKEN' \
    -d '{
      "name": "Updated Product Name",
      "description": "Updated description",
      "price": 39.99,
      "inventory": 150
    }'
    ``` 
    
3.  **Delete Product** (Admin Only / via a REST endpoint)
    
    
    ```
    curl -X DELETE http://localhost:8082/products/{:product_id} \
    -H 'Content-Type: application/json' \
    -H 'Authorization: Bearer YOUR_ADMIN_TOKEN'
    ``` 
    

### Order Mutations

1.  **Place Order**

    
    ```
    mutation {
      placeOrder(input: {
        items: [
          { product_id: 1, quantity: 20 }
          { product_id: 2, quantity: 1 }
        ]
      }) {
        message
        order_id
      }
    }
    ``` 
    

### Example Headers for GraphQL Playground

In the GraphQL Playground, set the headers like this for authenticated requests:

`{
  "Authorization": "Bearer YOUR_TOKEN"
}` 

Replace `YOUR_TOKEN` with the JWT token you received from the login endpoint.

## REST API Examples

**Register User**:

```
curl -X POST http://localhost:8081/register \
-H 'Content-Type: application/json' \
-d '{"username": "admin", "email": "admin@example.com", "password": "adminpass"}'
``` 

**Login User**:


```
curl -X POST http://localhost:8081/login \
-H 'Content-Type: application/json' \
-d '{"username": "admin", "password": "adminpass"}'
``` 

**Create Product**:


```
curl -X POST http://localhost:8082/products \
-H 'Content-Type: application/json' \
-H 'Authorization: Bearer YOUR_ADMIN_TOKEN' \
-d '{"name": "Product 12", "description": "A sample product 12", "price": 17.99, "inventory": 290}'
``` 

**Place Order**:


```
curl -X POST http://localhost:8083/orders \
-H 'Content-Type: application/json' \
-H 'Authorization: Bearer YOUR_USER_TOKEN' \
-d '{
  "items": [
    {"product_id": 1, "quantity": 20}
  ]
}'
``` 

**Update Product**:


```
curl -X PUT http://localhost:8082/products/{:product_id} \
-H 'Content-Type: application/json' \
-H 'Authorization: Bearer YOUR_ADMIN_TOKEN' \
-d '{
  "name": "Updated Product Name",
  "description": "Updated description",
  "price": 39.99,
  "inventory": 150
}'
``` 

Replace `YOUR_ADMIN_TOKEN` or `YOUR_USER_TOKEN` with the correct token obtained from the login endpoint.