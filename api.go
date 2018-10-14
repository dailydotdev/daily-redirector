package main

import (
	"fmt"
	"github.com/sony/gobreaker"
)

var apiUrl = getEnv("API_URL", "http://localhost:4000")
var cb = gobreaker.NewCircuitBreaker(gobreaker.Settings{Name: "API"})

type Post struct {
	Url string
}

var getPost = func(id string) (Post, error) {
	post, err := cb.Execute(func() (interface{}, error) {
		post := Post{}
		err := getJson(fmt.Sprintf("%s/v1/posts/%s", apiUrl, id), &post)
		if err != nil {
			return nil, err
		}

		return post, nil
	})
	return post.(Post), err
}
