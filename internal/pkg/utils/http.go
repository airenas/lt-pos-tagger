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

var (
	closedChan chan time.Time
	//ExpBackoffList list of backoff values for http retry delays
	ExpBackoffList = [...]int{0, 40, 80, 160, 320, 640, 1280}
)

func init() {
	closedChan = make(chan time.Time)
	close(closedChan)
}

//RandomWait returns wait timeout channel
// wait time id st ms randomized in interval [0.5, 1.5]*st
func RandomWait(st int) <-chan time.Time {
	if st <= 0 {
		return closedChan
	}
	return time.After(time.Duration(randNum(st)) * time.Millisecond)
}

func randNum(st int) float64 {
	// use full jitter select random from [0, st)
	// as noted in https://aws.amazon.com/blogs/architecture/exponential-backoff-and-jitter/
	return float64(st) * rand.Float64()
}
