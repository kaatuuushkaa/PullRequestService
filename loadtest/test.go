package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net/http"
	"os"
	"sort"
	"sync"
	"time"
)

type Result struct {
	Endpoint string
	Latency  time.Duration
	Status   int
	Err      error
}

type Config struct {
	BaseURL     string
	Duration    time.Duration
	Concurrency int
}

func main() {
	var (
		baseURL     = flag.String("base-url", "http://localhost:8080", "service base URL")
		durationStr = flag.String("duration", "30s", "test duration, e.g. 30s, 1m")
		concurrency = flag.Int("concurrency", 10, "number of concurrent workers")
	)
	flag.Parse()

	dur, err := time.ParseDuration(*durationStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid duration: %v\n", err)
		os.Exit(1)
	}

	cfg := Config{
		BaseURL:     *baseURL,
		Duration:    dur,
		Concurrency: *concurrency,
	}

	rand.Seed(time.Now().UnixNano())

	if err := run(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "load test failed: %v\n", err)
		os.Exit(1)
	}
}

func run(cfg Config) error {
	client := &http.Client{
		Timeout: 2 * time.Second,
	}

	if err := setupTeam(cfg.BaseURL, client); err != nil {
		return fmt.Errorf("setup team: %w", err)
	}

	fmt.Printf("Running load test: baseURL=%s, duration=%s, concurrency=%d\n",
		cfg.BaseURL, cfg.Duration, cfg.Concurrency)

	resultsCh := make(chan Result, 1000)
	var wg sync.WaitGroup

	start := time.Now()
	deadline := start.Add(cfg.Duration)

	for i := 0; i < cfg.Concurrency; i++ {
		wg.Add(1)
		go worker(i+1, cfg, deadline, client, resultsCh, &wg)
	}

	go func() {
		wg.Wait()
		close(resultsCh)
	}()

	var (
		allLatencies []time.Duration
		okCount      int
		errCount     int
		totalCount   int
		byEndpoint   = make(map[string][]time.Duration)
		statusCounts = make(map[int]int)
	)

	for res := range resultsCh {
		totalCount++
		if res.Status != 0 {
			statusCounts[res.Status]++
		}
		if res.Err != nil {
			errCount++
		} else {
			okCount++
			allLatencies = append(allLatencies, res.Latency)
			byEndpoint[res.Endpoint] = append(byEndpoint[res.Endpoint], res.Latency)
		}
	}

	totalDuration := time.Since(start)
	printSummary(totalCount, okCount, errCount, totalDuration, allLatencies, byEndpoint, statusCounts)

	return nil
}

func setupTeam(baseURL string, client *http.Client) error {
	teamPayload := map[string]any{
		"team_name": "payments",
		"members": []map[string]any{
			{"user_id": "u1", "username": "Alice", "is_active": true},
			{"user_id": "u2", "username": "Bob", "is_active": true},
			{"user_id": "u3", "username": "Carol", "is_active": true},
			{"user_id": "u4", "username": "Dave", "is_active": true},
			{"user_id": "u5", "username": "Eve", "is_active": true},
		},
	}

	body, err := json.Marshal(teamPayload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, baseURL+"/team/add", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusBadRequest {
		return fmt.Errorf("unexpected status creating team: %d", resp.StatusCode)
	}

	return nil
}

func worker(id int, cfg Config, deadline time.Time, client *http.Client, out chan<- Result, wg *sync.WaitGroup) {
	defer wg.Done()

	authors := []string{"u1", "u2", "u3", "u4", "u5"}
	iter := 0

	for time.Now().Before(deadline) {
		iter++
		author := authors[rand.Intn(len(authors))]
		prID := fmt.Sprintf("pr-%d-%d-%d", id, iter, time.Now().UnixNano())

		createReqBody := map[string]any{
			"pull_request_id":   prID,
			"pull_request_name": fmt.Sprintf("Load test PR %s", prID),
			"author_id":         author,
		}
		if res := doJSONRequest(client, cfg.BaseURL+"/pullRequest/create", createReqBody); res != nil {
			res.Endpoint = "POST /pullRequest/create"
			out <- *res
		}

		mergeReqBody := map[string]any{
			"pull_request_id": prID,
		}
		if res := doJSONRequest(client, cfg.BaseURL+"/pullRequest/merge", mergeReqBody); res != nil {
			res.Endpoint = "POST /pullRequest/merge"
			out <- *res
		}

		if iter%5 == 0 {
			userID := authors[rand.Intn(len(authors))]
			url := fmt.Sprintf("%s/users/getReview?user_id=%s", cfg.BaseURL, userID)
			if res := doGETRequest(client, url); res != nil {
				res.Endpoint = "GET /users/getReview"
				out <- *res
			}
		}

		if iter%10 == 0 {
			url := cfg.BaseURL + "/stats"
			if res := doGETRequest(client, url); res != nil {
				res.Endpoint = "GET /stats"
				out <- *res
			}
		}

		time.Sleep(100 * time.Millisecond)
	}
}

func doJSONRequest(client *http.Client, url string, payload any) *Result {
	body, err := json.Marshal(payload)
	if err != nil {
		return &Result{Err: fmt.Errorf("marshal: %w", err)}
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return &Result{Err: fmt.Errorf("request: %w", err)}
	}
	req.Header.Set("Content-Type", "application/json")

	start := time.Now()
	resp, err := client.Do(req)
	lat := time.Since(start)

	if err != nil {
		return &Result{Latency: lat, Err: err}
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)

	return &Result{
		Latency: lat,
		Status:  resp.StatusCode,
		Err:     nil,
	}
}

func doGETRequest(client *http.Client, url string) *Result {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return &Result{Err: fmt.Errorf("request: %w", err)}
	}

	start := time.Now()
	resp, err := client.Do(req)
	lat := time.Since(start)

	if err != nil {
		return &Result{Latency: lat, Err: err}
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)

	return &Result{
		Latency: lat,
		Status:  resp.StatusCode,
		Err:     nil,
	}
}

func printSummary(total, ok, failed int, duration time.Duration, latencies []time.Duration, byEndpoint map[string][]time.Duration, statusCounts map[int]int) {
	fmt.Println("----- Load test summary -----")
	fmt.Printf("Total requests: %d\n", total)
	fmt.Printf("Successful (no network error): %d\n", ok)
	fmt.Printf("Failed (network errors): %d\n", failed)
	if duration > 0 {
		rps := float64(total) / duration.Seconds()
		fmt.Printf("RPS: %.2f\n", rps)
	}

	if len(latencies) == 0 {
		fmt.Println("No latency data collected")
		return
	}

	sort.Slice(latencies, func(i, j int) bool { return latencies[i] < latencies[j] })
	fmt.Printf("Latency (overall):\n")
	fmt.Printf("  avg: %s\n", avgDuration(latencies))
	fmt.Printf("  p95: %s\n", percentile(latencies, 0.95))
	fmt.Printf("  p99: %s\n", percentile(latencies, 0.99))
	fmt.Printf("  max: %s\n", latencies[len(latencies)-1])

	fmt.Println("Latency by endpoint (avg, p95):")
	for ep, lats := range byEndpoint {
		sort.Slice(lats, func(i, j int) bool { return lats[i] < lats[j] })
		fmt.Printf("  %s: avg=%s p95=%s\n", ep, avgDuration(lats), percentile(lats, 0.95))
	}

	fmt.Println("Status codes:")
	for status, count := range statusCounts {
		fmt.Printf("  %d: %d\n", status, count)
	}
}

func avgDuration(durs []time.Duration) time.Duration {
	if len(durs) == 0 {
		return 0
	}
	var sum time.Duration
	for _, d := range durs {
		sum += d
	}
	return time.Duration(int64(sum) / int64(len(durs)))
}

func percentile(durs []time.Duration, p float64) time.Duration {
	if len(durs) == 0 {
		return 0
	}
	n := len(durs)
	index := int(math.Ceil(p*float64(n))) - 1
	if index < 0 {
		index = 0
	}
	if index >= n {
		index = n - 1
	}
	return durs[index]
}
