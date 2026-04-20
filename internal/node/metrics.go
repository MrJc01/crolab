// Copyright (c) 2026 Crolab Contributors. All rights reserved.
// Licensed under the Crolab Sustainable License (CSL).
// Contact: mrj.crom@gmail.com
package node

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
	"time"
)

// Metrics tracked by the node.
var (
	metricsJobsTotal     int64
	metricsJobsRunning   int64
	metricsJobsQueued    int64
	metricsJobsCompleted int64
	metricsJobsFailed    int64
	metricsStartTime     = time.Now()
)

func incTotal()     { atomic.AddInt64(&metricsJobsTotal, 1) }
func incRunning()   { atomic.AddInt64(&metricsJobsRunning, 1) }
func decRunning()   { atomic.AddInt64(&metricsJobsRunning, -1) }
func incQueued()    { atomic.AddInt64(&metricsJobsQueued, 1) }
func decQueued()    { atomic.AddInt64(&metricsJobsQueued, -1) }
func incCompleted() { atomic.AddInt64(&metricsJobsCompleted, 1) }
func incFailed()    { atomic.AddInt64(&metricsJobsFailed, 1) }

type MetricsSnapshot struct {
	Uptime       string `json:"uptime"`
	JobsTotal    int64  `json:"jobs_total"`
	JobsRunning  int64  `json:"jobs_running"`
	JobsQueued   int64  `json:"jobs_queued"`
	JobsComplete int64  `json:"jobs_completed"`
	JobsFailed   int64  `json:"jobs_failed"`
	MaxSlots     int    `json:"max_slots"`
	GPUs         []GPUInfo `json:"gpus"`
}

func getMetrics() MetricsSnapshot {
	return MetricsSnapshot{
		Uptime:       time.Since(metricsStartTime).Truncate(time.Second).String(),
		JobsTotal:    atomic.LoadInt64(&metricsJobsTotal),
		JobsRunning:  atomic.LoadInt64(&metricsJobsRunning),
		JobsQueued:   atomic.LoadInt64(&metricsJobsQueued),
		JobsComplete: atomic.LoadInt64(&metricsJobsCompleted),
		JobsFailed:   atomic.LoadInt64(&metricsJobsFailed),
		MaxSlots:     MaxConcurrentJobs,
		GPUs:         DetectGPUs(),
	}
}

// StartMetricsServer exposes /metrics as JSON on a separate HTTP port.
func StartMetricsServer(port string) {
	mux := http.NewServeMux()

	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(getMetrics())
	})

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status":"ok"}`)
	})

	log.Printf("📊 Métricas em http://localhost%s/metrics", port)
	go http.ListenAndServe(port, mux)
}
