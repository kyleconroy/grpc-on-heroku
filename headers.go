package main

import (
	"net/http"
	"strings"
	"time"
)

func Set(w http.ResponseWriter, hs ...Header) {
	for _, h := range hs {
		h.Populate(w.Header())
	}
}

type Header interface {
	Populate(http.Header)
}

type AccessControl struct {
	MaxAge           time.Duration
	Origin           string
	ExposedHeaders   []string
	AllowCredentials bool
	AllowedMethods   []string
	AllowedHeaders   []string
}

func (ac AccessControl) Populate(h http.Header) {
	if ac.Origin != "" {
		h.Set("Access-Control-Allow-Origin", ac.Origin)
	}
	if len(ac.AllowedMethods) > 0 {
		h.Set("Access-Control-Allow-Methods", strings.Join(ac.AllowedMethods, ", "))
	}
}

type ContentType string

func (ct ContentType) Populate(h http.Header) {
	h.Set("Content-Type", string(ct))
}
