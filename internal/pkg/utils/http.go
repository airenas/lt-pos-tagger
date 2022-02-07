package utils

import (
	"math/rand"
	"net/http"
	"time"
)

//IsRetryCode checks if http code indicates retryable error
func IsRetryCode(status int) bool {
	return status == http.StatusTooManyRequests || status == http.StatusServiceUnavailable
}

//RandomWait returns wait timeout channel
// wait time id st ms randomized in interval [0.5, 1.5]*st
func RandomWait(st int) <-chan time.Time {
	if st <= 0 {
		return time.After(time.Duration(0))
	}
	return time.After(time.Duration(randNum(st)) * time.Millisecond)
}

func randNum(st int) float64 {
	return float64(st) * (0.5 + rand.Float64())
}

//ExpBackoffList list of backoff values for http retry delays
var ExpBackoffList = [...]int{0, 40, 80, 160, 320, 640, 1280}
