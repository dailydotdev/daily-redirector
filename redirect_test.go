package main

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

var getPostSuccess = func(t *testing.T, expId string) func(id string, r *http.Request) (Post, error) {
	return func(id string, r *http.Request) (Post, error) {
		assert.Equal(t, id, expId, "wrong post id")

		return Post{
			Url: "https://www.dailynow.co",
		}, nil
	}
}

func TestRedirectOnBot(t *testing.T) {
	getPost = getPostSuccess(t, "post_id")

	req, err := http.NewRequest("GET", "/r/post_id", nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)")
	assert.Nil(t, err)

	rr := httptest.NewRecorder()

	router := createRouter()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusMovedPermanently, rr.Code, "wrong status code")

	assert.Equal(t, "https://www.dailynow.co", rr.Header().Get("Location"), "wrong redirect")
}

func TestRedirectOnBrowser(t *testing.T) {
	getPost = getPostSuccess(t, "post_id2")

	req, err := http.NewRequest("GET", "/r/post_id2", nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/60.0.3112.113 Safari/537.36")
	assert.Nil(t, err)

	rr := httptest.NewRecorder()

	router := createRouter()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code, "wrong status code")

	res := "<html><head><meta http-equiv=\"refresh\" content=\"0;URL=https://www.dailynow.co\"></head></html>\n"
	assert.Equal(t, res, rr.Body.String(), "wrong response")
}

func TestViewPublish(t *testing.T) {
	agent := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/60.0.3112.113 Safari/537.36"
	called := false

	getPost = getPostSuccess(t, "post_id3")

	publishView = func(view View) error {
		called = true
		assert.Equal(t, "post_id3", view.PostId)
		assert.Equal(t, "user_id", view.UserId)
		assert.Equal(t, agent, view.Agent)
		return nil
	}

	req, err := http.NewRequest("GET", "/r/post_id3", nil)
	req.Header.Set("User-Agent", agent)
	req.Header.Set("User-Id", "user_id")
	assert.Nil(t, err)

	rr := httptest.NewRecorder()

	router := createRouter()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code, "wrong status code")

	res := "<html><head><meta http-equiv=\"refresh\" content=\"0;URL=https://www.dailynow.co\"></head></html>\n"
	assert.Equal(t, rr.Body.String(), res, "wrong response")
	assert.Equal(t, true, called, "publishView should be called")
}

func TestAPIFail(t *testing.T) {
	getPost = func(id string, r *http.Request) (Post, error) {
		return Post{}, errors.New("fail")
	}

	req, err := http.NewRequest("GET", "/r/post_id4", nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/60.0.3112.113 Safari/537.36")
	assert.Nil(t, err)

	rr := httptest.NewRecorder()

	router := createRouter()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusServiceUnavailable, rr.Code, "wrong status code")
}

func TestNotFound(t *testing.T) {
	getPost = func(id string, r *http.Request) (Post, error) {
		return Post{}, errors.New("not found")
	}

	req, err := http.NewRequest("GET", "/r/post_id5", nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/60.0.3112.113 Safari/537.36")
	assert.Nil(t, err)

	rr := httptest.NewRecorder()

	router := createRouter()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code, "wrong status code")
}
