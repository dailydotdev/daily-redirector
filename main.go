package main

import (
	"fmt"
	"github.com/hashicorp/golang-lru"
	"github.com/julienschmidt/httprouter"
	"github.com/mssola/user_agent"
	log "github.com/sirupsen/logrus"
	"html/template"
	"net/http"
	"strconv"
	"time"
)

type RedirectData struct {
	Url string
}

type View struct {
	UserId    string
	PostId    string
	Referer   string
	Agent     string
	Ip        string
	Timestamp time.Time
}

var httpClient = &http.Client{Timeout: 5 * time.Second}
var cache *lru.Cache = nil

func RedirectBrowser(w http.ResponseWriter, r *http.Request, postId string, url string) {
	t, err := template.ParseFiles("tmpl/redirect.html")
	if err != nil {
		log.Panic("failed to parse redirect template ", err)
	}

	data := RedirectData{url}
	w.Header().Set("Cache-Control", "max-age=31536000")
	err = t.Execute(w, data)
	if err != nil {
		log.Error("failed to execute redirect template ", err)
		http.Error(w, "Server Internal Error", http.StatusInternalServerError)
	}

	userId := r.Header.Get("User-Id")
	if len(userId) > 0 {
		view := View{
			UserId:    userId,
			PostId:    postId,
			Referer:   r.Referer(),
			Agent:     r.UserAgent(),
			Ip:        r.RemoteAddr,
			Timestamp: time.Now(),
		}
		defer publishView(view)
	}
}

func Health(w http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
	fmt.Fprintf(w, "OK")
}

func Redirect(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	postId := ps.ByName("id")
	var post Post
	var err error

	if cache.Contains(postId) {
		if value, ok := cache.Get(postId); ok {
			post = value.(Post)
		} else {
			log.Error("failed to fetch from cache ")
			http.Error(w, "Server Internal Error", http.StatusInternalServerError)
			return
		}
	} else {
		log.Info("cache miss for post ", postId)
		post, err = getPost(postId)
		if err != nil {
			log.Warn("failed to get response from api ", err)
			http.Error(w, "Service Unavailable", http.StatusServiceUnavailable)
			return
		}
		cache.Add(postId, post)
	}

	ua := user_agent.New(r.UserAgent())
	if ua != nil && !ua.Bot() {
		RedirectBrowser(w, r, postId, post.Url)
	} else {
		http.Redirect(w, r, post.Url, http.StatusMovedPermanently)
	}
}

func createRouter() *httprouter.Router {
	router := httprouter.New()

	router.GET("/health", Health)
	router.GET("/r/:id", Redirect)
	return router
}

func init() {
	if getEnv("LOG_FORMAT", "text") == "json" {
		log.SetFormatter(&log.JSONFormatter{})
	}

	var err error
	pubsubClient, err = configurePubsub()
	if err != nil {
		log.Fatal("failed to initialize google pub/sub client ", err)
	}

	size, _ := strconv.Atoi(getEnv("CACHE_SIZE", "100"))
	cache, err = lru.New(size)
	if err != nil {
		log.Fatal("failed to initialize lru cache ", err)
	}
}

func main() {
	router := createRouter()
	addr := fmt.Sprintf(":%s", getEnv("PORT", "9090"))
	log.Info("server is listening to ", addr)
	err := http.ListenAndServe(addr, router) // set listen addr
	if err != nil {
		log.Fatal("failed to start listening ", err)
	}
}
