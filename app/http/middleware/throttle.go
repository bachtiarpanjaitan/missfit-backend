package middleware

import (
	"sync"
	"time"

	"github.com/goravel/framework/contracts/http"
)

type Visitor struct {
	Count     int
	LastVisit time.Time
}

var visitors = make(map[string]*Visitor)
var mu sync.Mutex

func Throttle(maxRequest int, duration time.Duration) http.Middleware {
	return func(ctx http.Context) {
		ip := ctx.Request().Ip()

		mu.Lock()

		visitor, exists := visitors[ip]

		if !exists {
			visitors[ip] = &Visitor{
				Count:     1,
				LastVisit: time.Now(),
			}

			mu.Unlock()
			ctx.Request().Next()
			return
		}

		// reset counter kalau sudah lewat durasi
		if time.Since(visitor.LastVisit) > duration {
			visitor.Count = 0
			visitor.LastVisit = time.Now()
		}

		visitor.Count++

		if visitor.Count > maxRequest {
			mu.Unlock()

			ctx.Response().Json(429, map[string]any{
				"message": "Terlalu banyak request. Coba lagi nanti.",
			})

			return
		}

		mu.Unlock()

		ctx.Request().Next()
	}
}
