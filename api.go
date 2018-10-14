package main

import (
	"fmt"
	"github.com/sony/gobreaker"
	"net/http"
)

var apiUrl = getEnv("API_URL", "http://localhost:4000")
var cb = gobreaker.NewCircuitBreaker(gobreaker.Settings{Name: "API"})

type Post struct {
	Url string
}

var getPost = func(id string, r *http.Request) (Post, error) {
	post, err := cb.Execute(func() (interface{}, error) {
		post := Post{}
		req, _ := http.NewRequest("GET", fmt.Sprintf("%s/v1/posts/%s", apiUrl, id), nil)
		req = req.WithContext(r.Context())
		err := getJson(req, &post)
		if err != nil {
			return nil, err
		}

		return post, nil
	})
	return post.(Post), err
}
