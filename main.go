package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
	"time"
)

type Device struct {
	AcceptedShares  int64   `json:"accepted_shares"`
	CoreClock       int64   `json:"core_clock"`
	CoreUtilization int64   `json:"core_utilization"`
	Fan             int64   `json:"fan"`
	Hashrate        string  `json:"hashrate"`
	Hashrate2       string  `json:"hashrate2"`
	HashrateRaw     float64 `json:"hashrate_raw"`
	Hashrate2Raw    float64 `json:"hashrate2_raw"`
	Id              int64   `json:"id"`
	InvalidShares   int64   `json:"invalid_shares"`
	MemClock        int64   `json:"mem_clock"`
	MemUtilization  int64   `json:"mem_utilization"`
	PCIBusID        int64   `json:"pci_bus_id"`
	Power           int64   `json:"power"`
	RejectedShares  int64   `json:"rejected_shares"`
	Temperature     int64   `json:"temperature"`
}

type MinerData struct {
	Devices           []Device
	TotalHashrate     string  `json:"total_hashrate"`
	TotalHashrate2    string  `json:"total_hashrate2"`
	TotalHashrate2Raw float64 `json:"total_hashrate2_raw"`
	TotalHashrateRaw  float64 `json:"total_hashrate_raw"`
	TotalPowerConsume int64   `json:"total_power_consume"`
}

type Stratum struct {
	AcceptedShares  int64 `json:"accepted_shares"`
	InvalidShares   int64 `json:"invalid_shares"`
	Latency         int64
	PoolHashrate10m string `json:"pool_hashrate_10m"`
	PoolHashrate24h string `json:"pool_hashrate_24h"`
	PoolHashrate4h  string `json:"pool_hashrate_4h"`
	RejectedShares  int64  `json:"rejected_shares"`
}

type NBMiner struct {
	RebootTimes int64 `json:"reboot_times"`
	StartTime   int64 `json:"start_time"`
	Stratum     Stratum
	MinerData   MinerData `json:"miner"`
}

var polling_error, parsing_error prometheus.Counter
var shares, invalid_shares, rejected_shares, latency, total_power prometheus.Gauge
var hostString, minerEndpoint string
var pollingInterval int

func poll(sleep func()) {
	defer sleep()
	resp, err := http.Get(fmt.Sprintf("%s/api/v1/status", minerEndpoint))
	if err != nil {
		polling_error.Inc()
		log.Print("Polling Error")
		return
	}
	defer resp.Body.Close()

	var data NBMiner
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		parsing_error.Inc()
		log.Print("Parsing Error", err)
		return
	}

	log.Printf("Shares: %d", data.Stratum.AcceptedShares)
	log.Printf("InvalidShares: %d", data.Stratum.InvalidShares)
	log.Printf("RejectedShares: %d", data.Stratum.RejectedShares)
	log.Printf("Latency: %d", data.Stratum.Latency)
	log.Printf("TotalPower: %d", data.MinerData.TotalPowerConsume)

	shares.Set(float64(data.Stratum.AcceptedShares))
	invalid_shares.Set(float64(data.Stratum.InvalidShares))
	rejected_shares.Set(float64(data.Stratum.RejectedShares))
	latency.Set(float64(data.Stratum.Latency))

	total_power.Set(float64(data.MinerData.TotalPowerConsume))

	log.Println("Polled")
}

func main() {
	flag.StringVar(&hostString, "host", ":2112", "host and port to bind to for which prometheus polls")
	flag.StringVar(&minerEndpoint, "minter", "http://localhost:22333", "the host and port where nbminer is exporting, {VALUE}/api/v1/status")
	flag.IntVar(&pollingInterval, "polling_interval", 30, "The number of seconds to sleep between polling invervals")

	flag.Parse()

	log.Printf("Host: %s", hostString)
	log.Printf("Miner: %s", minerEndpoint)
	log.Printf("Polling Seconds: %d", pollingInterval)

	polling_error = promauto.NewCounter(prometheus.CounterOpts{
		Name: "polling_errors",
		Help: "How many errors we've had polling nbminer",
	})
	parsing_error = promauto.NewCounter(prometheus.CounterOpts{
		Name: "parsing_errors",
		Help: "How many errors we've had parsing nbminer response",
	})

	shares = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "shares",
		Help: "Total number of shares processed",
	})
	invalid_shares = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "invalid_shares",
		Help: "Invalid shared processed",
	})
	rejected_shares = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "rejected_shares",
		Help: "Rejected shared processed",
	})
	latency = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "latency",
		Help: "Latency of publishing",
	})
	total_power = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "total_power",
		Help: "Total power in Watt Hours",
	})

	sleep := func() {
		time.Sleep(time.Second * time.Duration(pollingInterval))
	}

	// load the data first
	poll(func() {})

	// fire off an updater
	go func() {
		for {
			poll(sleep)
		}
	}()

	// start prom http
	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(hostString, nil)
}
