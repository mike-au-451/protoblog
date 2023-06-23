package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/rs/zerolog/log"
)

func hRoot(w http.ResponseWriter, r *http.Request) {
	rid := NewRequestId()
	log.Info().Msg(fmt.Sprintf("%s: hRoot: %s %s %s", rid, r.RemoteAddr, r.Method, r.URL))

	switch r.Method {
	case http.MethodGet:
		hGet(w, r, rid)
	case http.MethodPost:
		// unimplemented
		r405(w, r, rid, "not implemented")
	default:
		r405(w, r, rid, "")
	}
}

func hGet(w http.ResponseWriter, r *http.Request, rid string) {
	log.Trace().Msg(fmt.Sprintf("%s: hGet", rid))

	switch strings.ToLower(r.URL.Path) {
	case "/":
		hGetRoot(w, r, rid)
	case "/entries":
		hGetEntries(w, r, rid)
	default:
		body, ok := cc.Get(r.URL.Path[1:])
		if !ok {
			r404(w, r, rid, r.URL.Path)
			break
		}

		w.WriteHeader(http.StatusOK)
		w.Write(body)
	}
}

func hGetRoot(w http.ResponseWriter, r *http.Request, rid string) {
	log.Trace().Msg(fmt.Sprintf("%s: hGetRoot", rid))

	body, ok := cc.Get(www + "/" + "index.html")
	if !ok {
		r500(w, r, rid, "failed to get index")
	}

	w.WriteHeader(http.StatusOK)
	w.Write(body)
}

func hGetEntries(w http.ResponseWriter, r *http.Request, rid string) {
	log.Trace().Msg(fmt.Sprintf("%s: hGetEntries", rid))

	entries := db.GetEntries()
	if entries == nil {
		r500(w, r, rid, "")
		return
	}

	body, err := json.Marshal(entries)
	if err != nil {
		log.Printf("%s: hGetEntries: failed to marshal: %s", rid, err)
		r500(w, r, rid, "")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(body)
}

func r400(w http.ResponseWriter, r *http.Request, rid, msg string) {
	log.Info().Msg(fmt.Sprintf("%s: r400: %s", rid, msg))
	w.WriteHeader(http.StatusBadRequest)
	fmt.Fprintf(w, "Bad Request, quote %s", rid)
}

func r404(w http.ResponseWriter, r *http.Request, rid, msg string) {
	log.Info().Msg(fmt.Sprintf("%s: r404: %s", rid, msg))
	w.WriteHeader(http.StatusNotFound)
	fmt.Fprintf(w, "Not Found, quote %s", rid)
}

func r405(w http.ResponseWriter, r *http.Request, rid, msg string) {
	log.Info().Msg(fmt.Sprintf("%s: r405: %s", rid, msg))
	w.WriteHeader(http.StatusMethodNotAllowed)
	fmt.Fprintf(w, "Method Not Allowed, quote %s", rid)
}

func r500(w http.ResponseWriter, r *http.Request, rid, msg string) {
	log.Info().Msg(fmt.Sprintf("%s: r500: %s", rid, msg))
	w.WriteHeader(http.StatusInternalServerError)
	fmt.Fprintf(w, "Internal Server Error, quote %s", rid)
}

var requestId int

func NewRequestId() string {
	requestId++
	return fmt.Sprintf("%04d", requestId)
}