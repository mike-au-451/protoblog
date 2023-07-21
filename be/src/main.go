package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"

	"main/blogdb"
	"main/cache"
	log "main/logger"
	"main/render"
)

var (
	cc *cache.Cache
	bdb *blogdb.DB
)

func main() {
	if !environment() {
		log.Fatalf("failed to load environment")
	}

	hostport := os.Getenv("BLOG_HOST") + ":" + os.Getenv("BLOG_PORT")
	xx, _ := strconv.Atoi(os.Getenv("CACHE_SIZE"))
	cc = cache.New(os.Getenv("BLOG_PATH") + "/cache", xx)
	if cc == nil {
		log.Fatalf("failed to get cache")
	}

	bdb = blogdb.New(os.Getenv("DB_PATH"))
	if bdb == nil {
		log.Fatalf("failed to get db")
	}

	http.HandleFunc("/", hRoot)
	http.ListenAndServe(hostport, nil)

	bdb.Close()
}

func environment() bool {
	ok := true

	env := os.Getenv("ENV")
	if env == "" {
		env = "development"
	}
	err := godotenv.Load("env." + env)
	if err != nil {
		log.Errorf("failed to load environment %s, %s", "env." + env, err)
		ok = false
	}

	if os.Getenv("BLOG_HOST") == "" {
		log.Infof("defaulting BLOG_HOST")
		os.Setenv("BLOG_HOST", "localhost")
	}
	if os.Getenv("BLOG_PORT") == "" {
		log.Infof("defaulting BLOG_PORT")
		os.Setenv("BLOG_PORT", "8080")
	}
	if os.Getenv("BLOG_PATH") == "" {
		log.Errorf("environment missing BLOG_PATH")
		ok = false
	}
	if os.Getenv("ASSET_PATH") == "" {
		log.Errorf("environment missing ASSET_PATH")
		ok = false
	}
	if os.Getenv("CACHE_SIZE") == "" {
		log.Infof("defaulting CACHE_SIZE")
		os.Setenv("CACHE_SIZE", "10")
	}
	if os.Getenv("DB_PATH") == "" {
		log.Errorf("environment missing DB_PATH")
		ok = false
	}

	log.Infof("BLOG_HOST %s", os.Getenv("BLOG_HOST"))
	log.Infof("BLOG_PORT %s", os.Getenv("BLOG_PORT"))
	log.Infof("BLOG_PATH %s", os.Getenv("BLOG_PATH"))
	log.Infof("DB_PATH %s", os.Getenv("DB_PATH"))
	log.Infof("ASSET_PATH %s", os.Getenv("ASSET_PATH"))
	log.Infof("CACHE_SIZE %s", os.Getenv("CACHE_SIZE"))

	return ok
}

func hRoot(w http.ResponseWriter, r *http.Request) {
	rid := NewRequestId()
	ctx := context.WithValue(context.Background(), "rid", rid)
	log.Infof("%s hRoot: %s %s %s", rid, r.RemoteAddr, r.Method, r.URL.Path, )

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
		log.Errorf("failed to get entries")
		return
	}

	md := render.New()
	for ii := range entries {
		key := entries[ii].Hash
		content := cc.Get(key)
		if cc.Updated(key) {
			entries[ii].Body = string(content)
		} else {
			bb := bytes.Buffer{}
			err := md.Convert(content, &bb)
			if err != nil {
				log.Errorf("failed to render maekdown: %s", err)
				r500(ctx, w, r, "markdown render error")
				return
			}
			cc.Update(key, bb.Bytes())
			entries[ii].Body = bb.String()
		}
	}

	body, err := json.Marshal(entries)
	if err != nil {
		log.Errorf("failed to marshal: %s", err)
		r500(ctx, w, r, "json.Marshal error")
		return
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
	log.Errorf("%s r404: %s", ctx.Value("rid"), msg)

	w.Header().Set("Content-type", "text/html")
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte{})
}

func r405(ctx context.Context, w http.ResponseWriter, r *http.Request, msg string) {
	log.Errorf("%s r405: %s", ctx.Value("rid"), msg)

	w.Header().Set("Content-type", "text/html")
	w.WriteHeader(http.StatusMethodNotAllowed)
	w.Write([]byte{})
}

func r500(ctx context.Context, w http.ResponseWriter, r *http.Request, msg string) {
	log.Errorf("%s r500: %s", ctx.Value("rid"), msg)

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