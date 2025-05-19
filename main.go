package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/opensearch-project/opensearch-go"
	"github.com/opensearch-project/opensearch-go/opensearchapi"
)

var client opensearch.Client

func main() {
	r := gin.Default()

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	r.GET("/health", func(ctx *gin.Context) {
		client, err := createOpenSearchClient()
		if err != nil {
			log.Println("err while creating client", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"message": "error creating client",
			})
		}

		resp, err := getClusterHealth(client)
		if err != nil {
			log.Println("err getting cluster health", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"message": "error getting cluster health",
			})
		}
		ctx.JSON(http.StatusOK, gin.H{
			"message": resp,
		})
	})

	r.GET("/cluster-health", func(ctx *gin.Context) {
		client, err := createOpenSearchClient()
		if err != nil {
			log.Println("err while creating client", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"message": "error creating client",
			})
		}

		resp, err := getClusterHealthRawJSON(client)
		if err != nil {
			log.Println("err getting cluster health", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"message": "error getting cluster health",
			})
		}
		ctx.JSON(http.StatusOK, gin.H{
			"message": resp,
		})
	})

	r.Run()
}

func createOpenSearchClient() (*opensearch.Client, error) {
	client, err := opensearch.NewClient(opensearch.Config{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
		Addresses: []string{"https://opensearch-node1:9200"},
		Username:  "admin", // For testing only. Don't store credentials in code.
		Password:  "Jahnavi@21",
	})
	if err != nil {
		return nil, err
	}

	return client, nil
}

// Struct to unmarshal the cluster health response
type ClusterHealthResponse struct {
	Status string `json:"status"`
}

func getClusterHealth(client *opensearch.Client) (string, error) {
	req := opensearchapi.ClusterHealthRequest{}
	res, err := req.Do(context.Background(), client)
	if err != nil {
		log.Printf("Error getting response: %s", err)
		return "", err
	}
	defer res.Body.Close()

	var health ClusterHealthResponse
	if err := json.NewDecoder(res.Body).Decode(&health); err != nil {
		return "", fmt.Errorf("error parsing cluster health response: %w", err)
	}

	return health.Status, nil
}

func getClusterHealthRawJSON(client *opensearch.Client) (string, error) {
	req := opensearchapi.ClusterHealthRequest{}
	res, err := req.Do(context.Background(), client)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	if res.IsError() {
		return "", fmt.Errorf("error response from OpenSearch: %s", res.String())
	}

	var data map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&data); err != nil {
		return "", err
	}

	prettyJSON, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", err
	}

	log.Printf("Cluster Health JSON:\n%s", string(prettyJSON))
	return string(prettyJSON), nil
}
