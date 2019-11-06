package main

import (
	"flag"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/Pallinder/go-randomdata"
	wr "github.com/mroth/weightedrand"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const NginxTimeFormat = "02/Jan/2006:15:04:05 -0700"

type LogGen interface {
	String() (string, float64)
}

type NginxLog struct {
	Remote            string
	Host              string
	User              string
	Time              time.Time
	Method            string
	Path              string
	Code              int
	Size              int
	Referer           string
	Agent             string
	HttpXForwardedFor string
}

func NewNginxLog() NginxLog {
	return NginxLog{
		Remote:            "127.0.0.1",
		Host:              "-",
		User:              "-",
		Time:              time.Time{},
		Method:            "GET",
		Path:              "/loggen",
		Code:              200,
		Size:              650,
		Referer:           "-",
		Agent:             "golang/generator",
		HttpXForwardedFor: "-",
	}
}

func NewNginxLogRandom() NginxLog {
	rand.Seed(time.Now().UTC().UnixNano())
	c := wr.NewChooser(
		wr.Choice{Item: 200, Weight: 7},
		wr.Choice{Item: 404, Weight: 3},
		wr.Choice{Item: 503, Weight: 1},
		wr.Choice{Item: 302, Weight: 2},
		wr.Choice{Item: 403, Weight: 2},
	)
	return NginxLog{
		Remote:            randomdata.IpV4Address(),
		Host:              "-",
		User:              "-",
		Time:              time.Now(),
		Method:            randomdata.StringSample("GET", "POST", "PUT"),
		Path:              randomdata.StringSample("/", "/blog", "/index.html", "/products"),
		Code:              c.Pick().(int),
		Size:              rand.Intn(25000-100) + 100,
		Referer:           "-",
		Agent:             randomdata.UserAgentString(),
		HttpXForwardedFor: "-",
	}
}

func (n NginxLog) String() (string, float64) {
	message := fmt.Sprintf("%s %s %s [%s] \"%s %s HTTP/1.1\" %d %d %q %q %q", n.Remote, n.Host, n.User, n.Time.Format(NginxTimeFormat), n.Method, n.Path, n.Code, n.Size, n.Referer, n.Agent, n.HttpXForwardedFor)
	return message, float64(len([]byte(message)))
}

var (
	eventEmitted = promauto.NewCounter(prometheus.CounterOpts{
		Name: "event_emitted_total",
		Help: "The total number of events",
	})
	eventEmittedBytes = promauto.NewCounter(prometheus.CounterOpts{
		Name: "event_emitted_total_bytes",
		Help: "The total number of events",
	})
)

func main() {
	minIntervall := flag.String("min-intervall", "100ms", "Minimum interval between log messages")
	maxIntervall := flag.String("max-intervall", "1s", "Maximum interval between log messages")
	count := flag.Int("count", 0, "The amount of log message to emit.")
	randomise := flag.Bool("randomise", true, "Randomise log content")
	eventPerSec := flag.Int("event-per-sec", 10, "The amount of log message to emit/s")
	metricsAddr := flag.String("metrics.addr", ":11000", "Metrics server listen address")

	flag.Parse()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	done := make(chan bool)

	go func() {
		i := 0
		for {
			// If we have count check counter
			if *count > 0 && !(i < *count) {
				break
			}

			var n LogGen
			if *randomise {
				n = NewNginxLogRandom()
			} else {
				n = NewNginxLog()
			}

			msg, size := n.String()
			fmt.Println(msg)
			eventEmitted.Inc()
			eventEmittedBytes.Add(size)

			var duration time.Duration
			if *eventPerSec > 0 {
				sleepTime := 1000.0 / *eventPerSec
				duration = time.Duration(sleepTime) * time.Millisecond
			} else {
				// Randomise output between min and max
				parsedMaxTime, err := time.ParseDuration(*maxIntervall)
				if err != nil {
					panic(err)
				}
				parsedMinTime, err := time.ParseDuration(*minIntervall)
				if err != nil {
					panic(err)
				}
				duration = time.Duration(rand.Int63n(int64((parsedMaxTime - parsedMinTime) + parsedMinTime)))
			}
			// Sleep before next flush
			time.Sleep(duration)

			// Increment counter
			i++
		}
		done <- true
	}()

	go func() {
		fmt.Println("metrics listen on: ", *metricsAddr)
		http.Handle("/metrics", promhttp.Handler())
		http.ListenAndServe(*metricsAddr, nil)
		fmt.Println("main loop started")
	}()

	select {
	case <-done:
		return
	case <-interrupt:
		fmt.Println("Recieved interrupt")
		return
	}

}
