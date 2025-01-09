package main

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
)

func getAuthToken(client *http.Client) (token string, err error) {
	r, err := http.NewRequestWithContext(context.Background(), http.MethodPost, "http://localhost:8080/auth/login", strings.NewReader(`{"username":"ipesenti","password":"abcd1234!!"}`))
	if err != nil {
		return
	}
	r.Header.Set("Content-Type", "application/json")
	res, err := client.Do(r)
	if err != nil {
		return
	}
	defer res.Body.Close()
	data, err := io.ReadAll(res.Body)
	if err != nil {
		return
	}
	var tokenWrap struct {
		Token string `json:"token"`
	}
	err = json.Unmarshal(data, &tokenWrap)
	if err != nil {
		return
	}
	token = tokenWrap.Token
	return
}
