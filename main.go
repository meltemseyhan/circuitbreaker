package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/sony/gobreaker"
)

var cb *gobreaker.CircuitBreaker

func init() {
	var settings gobreaker.Settings
	settings.Name = "DUMMY SERVICE CIRCUIT BREAKER"
	{
		//MaxRequests is the maximum number of requests allowed to pass through when the CircuitBreaker is half-open.
		//If MaxRequests is 0, CircuitBreaker allows only 1 request.
		settings.MaxRequests = 3

		//Interval is the cyclic period of the closed state for CircuitBreaker to clear the internal Counts.
		//If Interval is 0, CircuitBreaker doesn't clear the internal Counts during the closed state.
		settings.Interval = time.Duration(30) * time.Minute

		//Timeout is the period of the open state, after which the state of CircuitBreaker becomes half-open.
		//If Timeout is 0, the timeout value of CircuitBreaker is set to 60 seconds.
		settings.Timeout = time.Duration(5) * time.Minute

		//ReadyToTrip is called with a copy of Counts whenever a request fails in the closed state.
		//If ReadyToTrip returns true, CircuitBreaker will be placed into the open state.
		//If ReadyToTrip is nil, default ReadyToTrip is used. Default ReadyToTrip returns true when the number of consecutive failures is more than 5.
		settings.ReadyToTrip = func(counts gobreaker.Counts) bool {
			failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
			return counts.Requests >= 8 && failureRatio >= 0.6
		}

		//OnStateChange is called whenever the state of CircuitBreaker changes.
		settings.OnStateChange = nil

		//IsSuccessful is called with the error returned from a request.
		//If IsSuccessful returns true, the error is counted as a success. Otherwise the error is counted as a failure.
		//If IsSuccessful is nil, default IsSuccessful is used, which returns false for all non-nil errors.
		settings.IsSuccessful = nil
	}
	cb = gobreaker.NewCircuitBreaker(settings)
}

func remoteCall() ([]byte, error) {
	resp, err := http.Get("http://localhost:8080/meltem")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

func main() {
	for i := 0; i < 100; i++ {
		/*
			b, err := remoteCall()
			if err != nil {
				log.Print(err)
			} else {
				fmt.Println(string(b))
			}
		*/
		b, err := cb.Execute(func() (interface{}, error) {
			return remoteCall()
		})
		if err != nil {
			log.Print(err)
		} else {
			fmt.Println(string(b.([]byte)))
		}
	}
}
