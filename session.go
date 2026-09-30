package main

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"sync"
	"time"
)

const sessionCookieName = "bn_wr_session"

type PlayerState struct {
	Found       map[string]bool // lead ID -> found
	HintsShown  map[string]int
	LoggedInAs  string
	DisplayName string
}

var (
	mu       sync.Mutex
	sessions = map[string]*PlayerState{}
)

func newID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func getOrCreate(w http.ResponseWriter, r *http.Request) *PlayerState {
	var id string
	if c, err := r.Cookie(sessionCookieName); err == nil && c.Value != "" {
		id = c.Value
	} else {
		id = newID()
		http.SetCookie(w, &http.Cookie{
			Name:     sessionCookieName,
			Value:    id,
			Path:     "/",
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
			Expires:  time.Now().Add(7 * 24 * time.Hour),
		})
	}
	mu.Lock()
	defer mu.Unlock()
	st, ok := sessions[id]
	if !ok {
		st = &PlayerState{Found: map[string]bool{}, HintsShown: map[string]int{}}
		sessions[id] = st
	}
	return st
}

func (p *PlayerState) foundCount() int {
	n := 0
	for _, v := range p.Found {
		if v {
			n++
		}
	}
	return n
}
