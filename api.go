package main

import (
	"errors"
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
		if err != nil && err.Error() == "404" {
			return nil, nil
		} else {
			return post, err
		}
	})

	if err != nil {
		return Post{}, err
	} else if post == nil {
		return Post{}, errors.New("not found")
	} else {
		return post.(Post), nil
	}
}
