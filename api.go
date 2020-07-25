package main

import (
	"errors"
	"fmt"
	"net/http"
)

var apiUrl = getEnv("API_URL", "http://localhost:4000")
var hystrixApi = "API"

type Post struct {
	Id  string
	Url string
}

var getPost = func(id string, r *http.Request) (Post, error) {
	post := Post{}
	req, _ := http.NewRequest("GET", fmt.Sprintf("%s/v1/posts/%s", apiUrl, id), nil)
	req = req.WithContext(r.Context())
	err := getJsonHystrix(hystrixApi, req, &post)
	if err != nil {
		if err.Error() == "404" {
			return Post{}, errors.New("not found")
		} else {
			return Post{}, err
		}
	}

	return post, nil
}
