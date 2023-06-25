package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"main/blogdb"
	"main/cache"
)

var (
	cc *cache.Cache
	db *blogdb.DB
	root, www string
)

func main() {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	zerolog.TimeFieldFormat = time.RFC3339Nano
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339Nano })
	log.Info().Msg("starting")

	env := os.Getenv("ENV")
	if env == "" {
		env = "development"
	}

	err := godotenv.Load("env." + env)
	if err != nil {
		log.Fatal().Msg(fmt.Sprintf("failed to load env: %s", err))
	}

	root = os.Getenv("BLOG_ROOT")
	sshDir := os.Getenv("SSH_DIR")
	dbDir := os.Getenv("DB_DIR")
	wwwDir := os.Getenv("WWW_DIR")
	srvr := os.Getenv("SRV_HOST")
	port := os.Getenv("SRV_PORT")
	sshCrt := os.Getenv("SSH_CERTIFICATE")
	sshKey := os.Getenv("SSH_KEY")
	cacheSize := os.Getenv("CACHE_SIZE")
	dbName := os.Getenv("DB_NAME")

	log.Info().Msg(fmt.Sprintf("BLOG_ROOT:       %s", root))
	log.Info().Msg(fmt.Sprintf("SSH_DIR:         %s", sshDir))
	log.Info().Msg(fmt.Sprintf("DB_DIR:          %s", dbDir))
	log.Info().Msg(fmt.Sprintf("WWW_DIR:         %s", wwwDir))
	log.Info().Msg(fmt.Sprintf("SRV_HOST:        %s", srvr))
	log.Info().Msg(fmt.Sprintf("SRV_PORT:        %s", port))
	log.Info().Msg(fmt.Sprintf("SSH_CERTIFICATE: %s", sshCrt))
	log.Info().Msg(fmt.Sprintf("SSH_KEY:         %s", sshKey))
	log.Info().Msg(fmt.Sprintf("CACHE_SIZE:      %s", cacheSize))
	log.Info().Msg(fmt.Sprintf("DB_NAME:         %s", dbName))

	xx, err := strconv.Atoi(cacheSize)
	if err != nil {
		log.Error().Msg(fmt.Sprintf("bad CACHE_SIZE: %s", err))
		xx = 10
	}
	cc = cache.New(root + "/" + wwwDir).SetSize(xx)

	db = blogdb.NewDB(root + "/" + dbDir + "/" + dbName, cc)
	if db == nil {
		log.Fatal().Msg(fmt.Sprintf("failed to get db at %s", root + "/" + dbDir + "/" + dbName))
	}

	if port != "" {
		srvr = srvr + ":" + port
	}

	http.HandleFunc("/", hRoot)

	log.Info().Msg(fmt.Sprintf("listening on %s", srvr))
	err = http.ListenAndServeTLS(srvr, root + "/" + sshDir + "/" + sshCrt, root + "/" + sshDir + "/" + sshKey, nil)
	if err != nil {
		log.Fatal().Msg(fmt.Sprintf("failed to serve: %s", err))
	}
}

