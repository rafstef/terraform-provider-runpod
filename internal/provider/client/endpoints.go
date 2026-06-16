package client

import "os"

func GetGraphQLEndpoint() string {
	endpoint := os.Getenv("RUNPOD_GRAPHQL_URL")
	if endpoint == "" {
		endpoint = "https://api.runpod.io/graphql"
	}
	return endpoint
}

func GetRestBaseURL() string {
	url := os.Getenv("RUNPOD_BASE_URL")
	if url == "" {
		url = "https://rest.runpod.io/v1"
	}
	return url
}
