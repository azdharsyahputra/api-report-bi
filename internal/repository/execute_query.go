package repository

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"os"
)

type queryRequest struct {
	Qstr string `json:"qstr"`
}

func ExecuteQuery(ctx context.Context, query string) (*http.Response, error) {

	body, _ := json.Marshal(queryRequest{
		Qstr: query,
	})

	req, err := http.NewRequestWithContext(
		ctx,
		"POST",
		os.Getenv("QUERY_SERVICE_URL"),
		bytes.NewBuffer(body),
	)

	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}

	return client.Do(req)
}
