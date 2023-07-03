package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	
	"main/cache"
	"main/blogdb"
	log "main/logger"
)

var (
	cc *cache.Cache
	bdb *blogdb.DB
)

func main() {
	env := os.Getenv("ENV")
	if env == "" {
		env = "development"
	}
	err := godotenv.Load("env." + env)
	if err != nil {
		log.Fatal("failed to load environment %s, %s", "env." + env, err)
	}

	blogHost := os.Getenv("BLOG_HOST")
	blogPort := os.Getenv("BLOG_PORT")
	blogRoot := os.Getenv("BLOG_ROOT")
	cacheDir := os.Getenv("CACHE_DIR")
	cacheSize := os.Getenv("CACHE_SIZE")
	dbDir := os.Getenv("DB_DIR")
	dbName := os.Getenv("DB_NAME")

	log.Info("BLOG_HOST %s", blogHost)
	log.Info("BLOG_PORT %s", blogPort)
	log.Info("BLOG_ROOT %s", blogRoot)
	log.Info("CACHE_DIR %s", cacheDir)
	log.Info("CACHE_SIZE %s", cacheSize)
	log.Info("DB_DIR %s", dbDir)
	log.Info("DB_NAME %s", dbName)

	if blogHost == "" {
		log.Fatal("missing BLOG_HOST in %s", "env." + env)
	}

	hostport := blogHost
	if blogPort != "" {
		hostport += ":" + blogPort
	}

	xx, _ := strconv.Atoi(cacheSize)
	cc = cache.New(blogRoot + "/" + cacheDir, xx)
	if cc == nil {
		log.Fatal("failed to get cache")
	}

	bdb = blogdb.New(blogRoot + "/" + dbDir + "/" + dbName)
	if bdb == nil {
		log.Fatal("failed to get db")
	}

	http.HandleFunc("/", hRoot)
	http.ListenAndServe(hostport, nil)

	bdb.Close()
}

func hRoot(w http.ResponseWriter, r *http.Request) {
	rid := NewRequestId()
	ctx := context.WithValue(context.Background(), "rid", rid)
	log.Info("%s hRoot", rid)

	switch r.Method {
	case http.MethodGet:
		hGet(ctx, w, r)
	default:
		r405(ctx, w, r, r.Method)
	}
}

func hGet(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	switch strings.ToLower(r.URL.Path) {
	case "/entries":
		hGetEntries(ctx, w, r)
	default:
		r404(ctx, w, r, r.URL.Path)
	}
}

func hGetEntries(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	entries := bdb.Entries()
	if entries == nil {
		log.Error("failed to get entries")
		return
	}

	for ii := range entries {
		entries[ii].Body = string(cc.Get(entries[ii].Body))
	}

	body, err := json.Marshal(entries)
	if err != nil {
		log.Error("failed to marshal: %s", err)
		r500(ctx, w, r, "json.Marshal error")
	}

	w.Header().Set("Content-type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusOK)
	w.Write(body)
}

const errorFormat = `
<!DOCTYPE html>
<html>
<head>
	<title>%s</title>
</head>
<body>
	<h1>%s</h1>
	<p>%s</p>
	<p>Quote error %s</p>
</body>
</html>
`

func r404(ctx context.Context, w http.ResponseWriter, r *http.Request, msg string) {
	log.Error("%s r404: %s", ctx.Value("rid"), msg)

	w.Header().Set("Content-type", "text/html")
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte{})
}

func r405(ctx context.Context, w http.ResponseWriter, r *http.Request, msg string) {
	log.Error("%s r405: %s", ctx.Value("rid"), msg)

	w.Header().Set("Content-type", "text/html")
	w.WriteHeader(http.StatusMethodNotAllowed)
	w.Write([]byte{})
}

func r500(ctx context.Context, w http.ResponseWriter, r *http.Request, msg string) {
	log.Error("%s r500: %s", ctx.Value("rid"), msg)

	errMsg := "Internal Server Error"
	body := []byte(fmt.Sprintf(errorFormat, errMsg, errMsg, "", ctx.Value("rid")))
	w.Header().Set("Content-type", "text/html")
	w.WriteHeader(http.StatusInternalServerError)
	w.Write(body)
}

var requestId int

func NewRequestId() string {
	requestId++
	return fmt.Sprintf("%04d", requestId)
}