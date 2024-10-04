package graph

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"graphql-gateway/graph/model"
	"io"
	"net/http"
	"os"
)

type Resolver struct{}

// Query Resolver Implementation
func (r *Resolver) Users(ctx context.Context) ([]*model.User, error) {
	url := fmt.Sprintf("%s/users", getUserServiceURL())

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", getAuthHeader(ctx))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to retrieve users: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result map[string][]map[string]interface{}
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}

	users, ok := result["users"]
	if !ok {
		return nil, fmt.Errorf("unexpected response format")
	}

	var usersPtr []*model.User
	for _, user := range users {
		idFloat, ok := user["id"].(float64)
		if !ok {
			return nil, fmt.Errorf("unexpected type for id")
		}

		userModel := &model.User{
			ID:        fmt.Sprintf("%d", int(idFloat)),
			Username:  user["username"].(string),
			Email:     user["email"].(string),
			CreatedAt: user["created_at"].(string),
		}
		usersPtr = append(usersPtr, userModel)
	}

	return usersPtr, nil
}

func (r *Resolver) User(ctx context.Context, id string) (*model.User, error) {
	url := fmt.Sprintf("%s/users/%s", getUserServiceURL(), id)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", getAuthHeader(ctx))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to retrieve user: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Unmarshal the JSON response into a map to extract the user data
	var result map[string]map[string]interface{}
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}

	userData, ok := result["user"]
	if !ok {
		return nil, fmt.Errorf("unexpected response format")
	}

	// Convert `id` to a string if it's an integer in the JSON response
	idFloat, ok := userData["id"].(float64)
	if !ok {
		return nil, fmt.Errorf("unexpected type for id")
	}

	user := &model.User{
		ID:        fmt.Sprintf("%d", int(idFloat)),
		Username:  userData["username"].(string),
		Email:     userData["email"].(string),
		CreatedAt: userData["created_at"].(string),
	}

	return user, nil
}

func (r *Resolver) RegisterUser(ctx context.Context, input model.RegisterInput) (*model.RegisterUserResponse, error) {
	url := fmt.Sprintf("%s/register", getUserServiceURL())

	userData, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(userData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to register user: %s", resp.Status)
	}

	var apiResponse map[string]string
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(body, &apiResponse)
	if err != nil {
		return nil, err
	}

	message, exists := apiResponse["message"]
	if !exists {
		return nil, fmt.Errorf("unexpected response format")
	}

	return &model.RegisterUserResponse{Message: message}, nil
}

// Helper function to get authorization header from context
func getAuthHeader(ctx context.Context) string {
	authHeader, _ := ctx.Value("Authorization").(string)
	return authHeader
}

// Helper function to get User Service URL
func getUserServiceURL() string {
	url := os.Getenv("USER_SERVICE_URL")
	if url == "" {
		url = "http://user-service:8081"
	}
	return url
}
