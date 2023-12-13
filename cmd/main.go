package main

import (
	"golang.org/x/time/rate"
	"net/http"
	"strings"
	"sync"
	"time"
)

type Session struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

var requests = make(map[string]Session)
var mu sync.Mutex

func limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		ipAddress := getIpFromXForwardedForHeader(r)
		limiter := getRequests(ipAddress)
		if limiter.limiter.Allow() == false {
			http.Error(w, http.StatusText(429), http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func getRequests(ipAddress string) Session {
	mu.Lock()

	session, exists := requests[ipAddress]

	if !exists {
		limiter := rate.NewLimiter(3, 1)
		session = Session{limiter: limiter, lastSeen: time.Now()}

		requests[ipAddress] = session
	}

	mu.Unlock()

	return session
}

func clearSessions() {
	for {

		time.Sleep(10 * time.Second)

		mu.Lock()

		for ip, session := range requests {
			if time.Since(session.lastSeen) > 2*time.Minute {
				delete(requests, ip)
			}
		}
		mu.Unlock()
	}
}

func getIpFromXForwardedForHeader(r *http.Request) string {
	forwardedForHeader := r.Header.Get("X-Forwarded-For")

	if forwardedForHeader == "" {
		ips := strings.Split(forwardedForHeader, ", ")

		return ips[0]
	}

	return ""
}

func main() {

}
