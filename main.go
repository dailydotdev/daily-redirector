package main

import (
	"contrib.go.opencensus.io/exporter/stackdriver"
	"contrib.go.opencensus.io/exporter/stackdriver/propagation"
	"fmt"
	"github.com/hashicorp/golang-lru"
	"github.com/julienschmidt/httprouter"
	"github.com/mssola/user_agent"
	log "github.com/sirupsen/logrus"
	"go.opencensus.io/plugin/ochttp"
	"go.opencensus.io/trace"
	_ "go.uber.org/automaxprocs"
	"google.golang.org/api/option"
	"html/template"
	"net/http"
	"os"
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

var httpClient *http.Client
var cache *lru.Cache
var gcpOpts []option.ClientOption
var tmpl *template.Template

func RedirectBrowser(w http.ResponseWriter, r *http.Request, postId string, url string) {
	data := RedirectData{url}
	err := tmpl.Execute(w, data)
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

	if value, ok := cache.Get(postId); ok {
		post = value.(Post)
	} else {
		log.Info("cache miss for post ", postId)
		post, err = getPost(postId, r)
		if err != nil {
			if err.Error() == "not found" {
				http.NotFound(w, r)
			} else {
				log.Warn("failed to get response from api ", err)
				http.Error(w, "Service Unavailable", http.StatusServiceUnavailable)
			}
			return
		}
		defer cache.Add(postId, post)
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
	if file, ok := os.LookupEnv("GOOGLE_APPLICATION_CREDENTIALS"); ok {
		gcpOpts = append(gcpOpts, option.WithCredentialsFile(file))
	}

	log.SetOutput(os.Stdout)
	if getEnv("ENV", "DEV") == "PROD" {
		log.SetFormatter(&log.JSONFormatter{})

		exporter, err := stackdriver.NewExporter(stackdriver.Options{
			ProjectID:          os.Getenv("GCLOUD_PROJECT"),
			TraceClientOptions: gcpOpts,
		})
		if err != nil {
			log.Fatal(err)
		}
		log.Info("registering stackdriver tracer")
		trace.RegisterExporter(exporter)
		trace.ApplyConfig(trace.Config{DefaultSampler: trace.AlwaysSample()})

		httpClient = &http.Client{
			Timeout: 5 * time.Second,
			Transport: &ochttp.Transport{
				// Use Google Cloud propagation format.
				Propagation: &propagation.HTTPFormat{},
			},
		}
	} else {
		httpClient = &http.Client{Timeout: 5 * time.Second}
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

	tmpl, err = template.ParseFiles("tmpl/redirect.html")
	if err != nil {
		log.Fatal("failed to parse redirect template ", err)
	}
}

func main() {
	router := createRouter()
	addr := fmt.Sprintf(":%s", getEnv("PORT", "9090"))
	log.Info("server is listening to ", addr)
	err := http.ListenAndServe(addr, &ochttp.Handler{Handler: router, Propagation: &propagation.HTTPFormat{}}) // set listen addr
	if err != nil {
		log.Fatal("failed to start listening ", err)
	}
}
