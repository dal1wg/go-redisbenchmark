/*
 * go-redisbenchmark - A high performance Redis benchmark tool written in Go.
 * Supports Redis 6.0+ features including ACL and TLS.
 *
 * Copyright (c) 2025, dal1wg <15072697283@139.com>
 * All rights reserved.
 *
 * This source code is licensed under the BSD 3-Clause License.
 */

package benchmark

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"net/url"
	"os"
	"strconv"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	progressbar "github.com/schollz/progressbar/v3"
	"github.com/redis/go-redis/v9"
)

// Mode represents Redis deployment mode.
type Mode string

const (
	ModeStandalone Mode = "standalone"
	ModeCluster   Mode = "cluster"
	ModeSentinel  Mode = "sentinel"
)

// Options configures a benchmark run.
type Options struct {
	Mode           Mode

	// Connection
	URI            string // Redis URI (e.g., redis://username:password@host:port/db?tls=true)

	// Standalone
	Addr           string // host:port
	DB             int

	// Shared
	Password       string

	// ACL support (Redis 6.0+)
	Username       string // ACL username
	ACLPassword    string // ACL password (separate from legacy password)

	// TLS support (Redis 6.0+)
	UseTLS         bool   // Enable TLS connection
	TLSCertFile    string // Client certificate file path
	TLSKeyFile     string // Client private key file path
	TLSCAFile      string // CA certificate file path
	TLSInsecure   bool   // Skip TLS certificate verification

	// Cluster
	ClusterAddrs   []string // host:port list

	// Sentinel
	SentinelAddrs  []string // host:port list
	SentinelMaster string

	// Workload
	Requests       int // total number of requests per test
	Concurrency    int // parallel workers
	Pipeline       int // number of pipelined commands per round-trip
	DataSize       int // payload size for SET/GET
	Tests          []string // e.g., ["PING","SET","GET","HGET",...]
	Keyspace       int // number of distinct keys per type
	Randomize      bool // randomize keys each request
	Quiet          bool // suppress per-interval logs (still shows progress)

	// Dataset sizes for complex types
	HashFields     int // fields per hash
	ListLen        int // elements per list
	SetCardinality int // members per set
	ZSetCardinality int // members per zset
	BitmapSizeBits int // bitmap length in bits
	GeoMembers     int // members per geo set
	StreamLen      int // entries per stream
	RangeCount     int // count for range reads (LRANGE/ZRANGE/XRANGE)
	
	// Data lifecycle management
	TTL            time.Duration // TTL for benchmark data (default: 1 hour)
}

// Result contains aggregated metrics for a single test.
type Result struct {
	TestName        string
	Requests        int
	Concurrency     int
	Pipeline        int
	PayloadBytes    int
	Elapsed         time.Duration
	OpsPerSec       float64
	ThroughputReqPS float64

	// Latency distribution in microseconds
	P50us           float64
	P95us           float64
	P99us           float64
}

// Runner executes benchmarks based on Options.
type Runner struct {
	opts   Options
}

func NewRunner(opts Options) (*Runner, error) {
	if opts.Requests <= 0 {
		return nil, errors.New("requests must be > 0")
	}
	if opts.Concurrency <= 0 {
		return nil, errors.New("concurrency must be > 0")
	}
	if opts.Pipeline <= 0 {
		opts.Pipeline = 1
	}
	if opts.DataSize < 0 {
		opts.DataSize = 0
	}
	if len(opts.Tests) == 0 {
		opts.Tests = []string{"PING"}
	}
	if opts.Keyspace <= 0 {
		opts.Keyspace = 1000
	}
	// defaults for dataset sizes
	if opts.HashFields <= 0 { opts.HashFields = 10 }
	if opts.ListLen <= 0 { opts.ListLen = 10 }
	if opts.SetCardinality <= 0 { opts.SetCardinality = 10 }
	if opts.ZSetCardinality <= 0 { opts.ZSetCardinality = 10 }
	if opts.BitmapSizeBits <= 0 { opts.BitmapSizeBits = 1024 }
	if opts.GeoMembers <= 0 { opts.GeoMembers = 10 }
	if opts.StreamLen <= 0 { opts.StreamLen = 10 }
	if opts.RangeCount <= 0 { opts.RangeCount = 10 }
	// Set default TTL to 1 hour if not specified
	if opts.TTL <= 0 { opts.TTL = time.Minute }
	return &Runner{opts: opts}, nil
}

// ParseRedisURI parses a Redis URI and returns connection parameters
// Supported formats:
// - redis://username:password@host:port/db?tls=true
// - rediss://username:password@host:port/db (TLS by default)
// - redis://host:port/db
// - redis://host:port
func (r *Runner) ParseRedisURI(uri string) error {
	if uri == "" {
		return nil
	}

	parsedURL, err := url.Parse(uri)
	if err != nil {
		return fmt.Errorf("invalid Redis URI: %w", err)
	}

	// Check scheme
	if parsedURL.Scheme != "redis" && parsedURL.Scheme != "rediss" {
		return fmt.Errorf("unsupported scheme: %s (only redis:// and rediss:// are supported)", parsedURL.Scheme)
	}

	// Set TLS if scheme is rediss://
	if parsedURL.Scheme == "rediss" {
		r.opts.UseTLS = true
	}

	// Parse host and port
	if parsedURL.Host != "" {
		r.opts.Addr = parsedURL.Host
	}

	// Parse username and password
	if parsedURL.User != nil {
		r.opts.Username = parsedURL.User.Username()
		if password, ok := parsedURL.User.Password(); ok {
			r.opts.ACLPassword = password
		}
	}

	// Parse database number
	if parsedURL.Path != "" && parsedURL.Path != "/" {
		dbStr := strings.TrimPrefix(parsedURL.Path, "/")
		if dbStr != "" {
			db, err := strconv.Atoi(dbStr)
			if err != nil {
				return fmt.Errorf("invalid database number in URI: %s", dbStr)
			}
			r.opts.DB = db
		}
	}

	// Parse query parameters
	query := parsedURL.Query()
	
	// Parse TLS parameters
	if query.Get("tls") == "true" || r.opts.UseTLS {
		r.opts.UseTLS = true
	}
	
	if certFile := query.Get("tls-cert"); certFile != "" {
		r.opts.TLSCertFile = certFile
	}
	
	if keyFile := query.Get("tls-key"); keyFile != "" {
		r.opts.TLSKeyFile = keyFile
	}
	
	if caFile := query.Get("tls-ca"); caFile != "" {
		r.opts.TLSCAFile = caFile
	}
	
	if query.Get("tls-insecure") == "true" {
		r.opts.TLSInsecure = true
	}

	return nil
}

// Run executes all tests sequentially and returns results.
func (r *Runner) Run(ctx context.Context) ([]Result, error) {
	results := make([]Result, 0, len(r.opts.Tests))

	client, err := r.buildClient()
	if err != nil {
		return nil, err
	}
	defer func() { _ = client.Close() }()

	// Pre-fill dataset for all non-PING tests
	if err := r.prefillAll(ctx, client); err != nil {
		return nil, fmt.Errorf("prefill error: %w", err)
	}

	for _, t := range r.opts.Tests {
		res, err := r.runSingleTest(ctx, client, strings.ToUpper(strings.TrimSpace(t)))
		if err != nil {
			return results, err
		}
		results = append(results, res)
	}
	return results, nil
}

func (r *Runner) buildClient() (redis.UniversalClient, error) {
	// Parse Redis URI if provided
	if err := r.ParseRedisURI(r.opts.URI); err != nil {
		return nil, fmt.Errorf("failed to parse Redis URI: %w", err)
	}

	// Use UniversalClient to support standalone/cluster/sentinel via a single config.
	opts := &redis.UniversalOptions{
		DB:       r.opts.DB,
		Password: r.opts.Password,
	}

	// ACL support (Redis 6.0+)
	if r.opts.Username != "" {
		opts.Username = r.opts.Username
		// If ACL password is specified, use it; otherwise fall back to legacy password
		if r.opts.ACLPassword != "" {
			opts.Password = r.opts.ACLPassword
		}
		// Debug: print ACL credentials
		fmt.Printf("Using ACL authentication - Username: %s, Password: %s\n", r.opts.Username, strings.Repeat("*", len(r.opts.ACLPassword)))
	}

	// TLS support (Redis 6.0+)
	if r.opts.UseTLS {
		tlsConfig, err := r.buildTLSConfig()
		if err != nil {
			return nil, fmt.Errorf("failed to build TLS config: %w", err)
		}
		opts.TLSConfig = tlsConfig
	}

	switch r.opts.Mode {
	case ModeStandalone:
		if r.opts.Addr == "" {
			opts.Addrs = []string{"127.0.0.1:6379"}
		} else {
			opts.Addrs = []string{r.opts.Addr}
		}
	case ModeCluster:
		if len(r.opts.ClusterAddrs) == 0 {
			if r.opts.Addr != "" {
				opts.Addrs = []string{r.opts.Addr}
			} else {
				opts.Addrs = []string{"127.0.0.1:6379"}
			}
		} else {
			opts.Addrs = r.opts.ClusterAddrs
		}
	case ModeSentinel:
		if len(r.opts.SentinelAddrs) == 0 || r.opts.SentinelMaster == "" {
			return nil, errors.New("sentinel mode requires sentinel addrs and master name")
		}
		opts.Addrs = r.opts.SentinelAddrs
		opts.MasterName = r.opts.SentinelMaster
	default:
		return nil, fmt.Errorf("unknown mode: %s", r.opts.Mode)
	}
	return redis.NewUniversalClient(opts), nil
}

// buildTLSConfig creates a TLS configuration for Redis connections
func (r *Runner) buildTLSConfig() (*tls.Config, error) {
	config := &tls.Config{
		MinVersion: tls.VersionTLS12,
	}

	// Load client certificate and key if provided
	if r.opts.TLSCertFile != "" && r.opts.TLSKeyFile != "" {
		cert, err := tls.LoadX509KeyPair(r.opts.TLSCertFile, r.opts.TLSKeyFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load client certificate: %w", err)
		}
		config.Certificates = []tls.Certificate{cert}
	}

	// Load CA certificate if provided
	if r.opts.TLSCAFile != "" {
		caCert, err := os.ReadFile(r.opts.TLSCAFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read CA certificate: %w", err)
		}

		caCertPool := x509.NewCertPool()
		if !caCertPool.AppendCertsFromPEM(caCert) {
			return nil, errors.New("failed to append CA certificate to pool")
		}
		config.RootCAs = caCertPool
	}

	// Set insecure skip verify if requested
	if r.opts.TLSInsecure {
		config.InsecureSkipVerify = true
	}

	return config, nil
}

// prefillAll creates datasets for reads across Redis types so read tests hit existing data.
func (r *Runner) prefillAll(ctx context.Context, client redis.UniversalClient) error {
	// strings for GET
	if containsTest(r.opts.Tests, "GET") {
		if err := r.prefillStrings(ctx, client); err != nil { return err }
	}
	// hashes
	if hasAny(r.opts.Tests, "HGET", "HGETALL", "HMGET") {
		if err := r.prefillHashes(ctx, client); err != nil { return err }
	}
	// lists
	if hasAny(r.opts.Tests, "LINDEX", "LRANGE_", "LLEN") {
		if err := r.prefillLists(ctx, client); err != nil { return err }
	}
	// sets
	if hasAny(r.opts.Tests, "SISMEMBER", "SMEMBERS", "SCARD") {
		if err := r.prefillSets(ctx, client); err != nil { return err }
	}
	// zsets
	if hasAny(r.opts.Tests, "ZSCORE", "ZRANGE_", "ZCARD", "ZRANK") {
		if err := r.prefillZSets(ctx, client); err != nil { return err }
	}
	// bitmaps
	if hasAny(r.opts.Tests, "GETBIT", "BITCOUNT") {
		if err := r.prefillBitmaps(ctx, client); err != nil { return err }
	}
	// hyperloglog
	if containsTest(r.opts.Tests, "PFCOUNT") {
		if err := r.prefillHLL(ctx, client); err != nil { return err }
	}
	// geo
	if hasAny(r.opts.Tests, "GEOPOS", "GEODIST") {
		if err := r.prefillGeo(ctx, client); err != nil { return err }
	}
	// streams
	if hasAny(r.opts.Tests, "XRANGE_", "XINFO", "XLEN") {
		if err := r.prefillStreams(ctx, client); err != nil { return err }
	}
	return nil
}

func (r *Runner) prefillStrings(ctx context.Context, client redis.UniversalClient) error {
	payload := randomString(max(1, r.opts.DataSize))
	pipe := client.Pipeline()
	count := 0
	// Use configured TTL for data cleanup
	ttl := r.opts.TTL
	for i := 0; i < r.opts.Keyspace; i++ {
		pipe.Set(ctx, r.keyS(i), payload, ttl)
		count++
		if count >= 256 {
			if _, err := pipe.Exec(ctx); err != nil && !isNilErr(err) { return err }
			pipe = client.Pipeline(); count = 0
		}
	}
	if count > 0 {
		if _, err := pipe.Exec(ctx); err != nil && !isNilErr(err) { return err }
	}
	return nil
}

func (r *Runner) prefillHashes(ctx context.Context, client redis.UniversalClient) error {
	pipe := client.Pipeline()
	count := 0
	// Use configured TTL for data cleanup
	ttl := r.opts.TTL
	for i := 0; i < r.opts.Keyspace; i++ {
		key := r.keyH(i)
		for f := 0; f < r.opts.HashFields; f++ {
			pipe.HSet(ctx, key, fmt.Sprintf("f%d", f), randomString(8))
			count++
			if count >= 512 { if _, err := pipe.Exec(ctx); err != nil && !isNilErr(err) { return err }; pipe = client.Pipeline(); count = 0 }
		}
		// Set TTL for the hash key after all fields are added
		pipe.Expire(ctx, key, ttl)
	}
	if count > 0 { if _, err := pipe.Exec(ctx); err != nil && !isNilErr(err) { return err } }
	return nil
}

func (r *Runner) prefillLists(ctx context.Context, client redis.UniversalClient) error {
	pipe := client.Pipeline()
	count := 0
	// Use configured TTL for data cleanup
	ttl := r.opts.TTL
	for i := 0; i < r.opts.Keyspace; i++ {
		key := r.keyL(i)
		for j := 0; j < r.opts.ListLen; j++ {
			pipe.RPush(ctx, key, fmt.Sprintf("e%d", j))
			count++
			if count >= 512 { if _, err := pipe.Exec(ctx); err != nil && !isNilErr(err) { return err }; pipe = client.Pipeline(); count = 0 }
		}
		// Set TTL for the list key after all elements are added
		pipe.Expire(ctx, key, ttl)
	}
	if count > 0 { if _, err := pipe.Exec(ctx); err != nil && !isNilErr(err) { return err } }
	return nil
}

func (r *Runner) prefillSets(ctx context.Context, client redis.UniversalClient) error {
	pipe := client.Pipeline()
	count := 0
	// Use configured TTL for data cleanup
	ttl := r.opts.TTL
	for i := 0; i < r.opts.Keyspace; i++ {
		key := r.keySet(i)
		for j := 0; j < r.opts.SetCardinality; j++ {
			pipe.SAdd(ctx, key, fmt.Sprintf("m%d", j))
			count++
			if count >= 512 { if _, err := pipe.Exec(ctx); err != nil && !isNilErr(err) { return err }; pipe = client.Pipeline(); count = 0 }
		}
		// Set TTL for the set key after all members are added
		pipe.Expire(ctx, key, ttl)
	}
	if count > 0 { if _, err := pipe.Exec(ctx); err != nil && !isNilErr(err) { return err } }
	return nil
}

func (r *Runner) prefillZSets(ctx context.Context, client redis.UniversalClient) error {
	pipe := client.Pipeline()
	count := 0
	// Use configured TTL for data cleanup
	ttl := r.opts.TTL
	for i := 0; i < r.opts.Keyspace; i++ {
		key := r.keyZ(i)
		for j := 0; j < r.opts.ZSetCardinality; j++ {
			pipe.ZAdd(ctx, key, redis.Z{Score: float64(j), Member: fmt.Sprintf("z%d", j)})
			count++
			if count >= 256 { if _, err := pipe.Exec(ctx); err != nil && !isNilErr(err) { return err }; pipe = client.Pipeline(); count = 0 }
		}
		// Set TTL for the zset key after all members are added
		pipe.Expire(ctx, key, ttl)
	}
	if count > 0 { if _, err := pipe.Exec(ctx); err != nil && !isNilErr(err) { return err } }
	return nil
}

func (r *Runner) prefillBitmaps(ctx context.Context, client redis.UniversalClient) error {
	pipe := client.Pipeline()
	count := 0
	// Use configured TTL for data cleanup
	ttl := r.opts.TTL
	for i := 0; i < r.opts.Keyspace; i++ {
		key := r.keyB(i)
		for bit := 0; bit < r.opts.BitmapSizeBits; bit += 3 {
			pipe.SetBit(ctx, key, int64(bit), 1)
			count++
			if count >= 1024 { if _, err := pipe.Exec(ctx); err != nil && !isNilErr(err) { return err }; pipe = client.Pipeline(); count = 0 }
		}
		// Set TTL for the bitmap key after all bits are set
		pipe.Expire(ctx, key, ttl)
	}
	if count > 0 { if _, err := pipe.Exec(ctx); err != nil && !isNilErr(err) { return err } }
	return nil
}

func (r *Runner) prefillHLL(ctx context.Context, client redis.UniversalClient) error {
	pipe := client.Pipeline()
	count := 0
	// Use configured TTL for data cleanup
	ttl := r.opts.TTL
	for i := 0; i < r.opts.Keyspace; i++ {
		key := r.keyPF(i)
		for j := 0; j < 64; j++ {
			pipe.PFAdd(ctx, key, fmt.Sprintf("hll%d-%d", i, j))
			count++
			if count >= 512 { if _, err := pipe.Exec(ctx); err != nil && !isNilErr(err) { return err }; pipe = client.Pipeline(); count = 0 }
		}
		// Set TTL for the hyperloglog key after all members are added
		pipe.Expire(ctx, key, ttl)
	}
	if count > 0 { if _, err := pipe.Exec(ctx); err != nil && !isNilErr(err) { return err } }
	return nil
}

func (r *Runner) prefillGeo(ctx context.Context, client redis.UniversalClient) error {
	pipe := client.Pipeline()
	count := 0
	// Use configured TTL for data cleanup
	ttl := r.opts.TTL
	for i := 0; i < r.opts.Keyspace; i++ {
		key := r.keyG(i)
		for j := 0; j < r.opts.GeoMembers; j++ {
			lon := -180.0 + float64((i+j)%360)
			lat := -85.0 + float64((i*2+j)%170)
			pipe.GeoAdd(ctx, key, &redis.GeoLocation{Name: fmt.Sprintf("g%d", j), Longitude: lon, Latitude: lat})
			count++
			if count >= 256 { if _, err := pipe.Exec(ctx); err != nil && !isNilErr(err) { return err }; pipe = client.Pipeline(); count = 0 }
		}
		// Set TTL for the geo key after all members are added
		pipe.Expire(ctx, key, ttl)
	}
	if count > 0 { if _, err := pipe.Exec(ctx); err != nil && !isNilErr(err) { return err } }
	return nil
}

func (r *Runner) prefillStreams(ctx context.Context, client redis.UniversalClient) error {
	pipe := client.Pipeline()
	count := 0
	// Use configured TTL for data cleanup
	ttl := r.opts.TTL
	for i := 0; i < r.opts.Keyspace; i++ {
		key := r.keyX(i)
		for j := 0; j < r.opts.StreamLen; j++ {
			pipe.XAdd(ctx, &redis.XAddArgs{Stream: key, Values: map[string]any{"f": fmt.Sprintf("v%d", j)}})
			count++
			if count >= 256 { if _, err := pipe.Exec(ctx); err != nil && !isNilErr(err) { return err }; pipe = client.Pipeline(); count = 0 }
		}
		// Set TTL for the stream key after all entries are added
		pipe.Expire(ctx, key, ttl)
	}
	if count > 0 { if _, err := pipe.Exec(ctx); err != nil && !isNilErr(err) { return err } }
	return nil
}

func (r *Runner) runSingleTest(ctx context.Context, client redis.UniversalClient, test string) (Result, error) {
	var (
		doneCount int64
		start    = time.Now()
		latCh    = make(chan time.Duration, 1024)
		wg       sync.WaitGroup
	)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// per-test progress bar
	var bar *progressbar.ProgressBar
	if !r.opts.Quiet {
		bar = progressbar.NewOptions64(
			int64(r.opts.Requests),
			progressbar.OptionSetDescription(fmt.Sprintf("%s", test)),
			progressbar.OptionSetWriter(os.Stdout),
			progressbar.OptionSetWidth(40),
			progressbar.OptionShowCount(),
			progressbar.OptionShowIts(),
			progressbar.OptionThrottle(65*time.Millisecond),
			progressbar.OptionSetTheme(progressbar.Theme{Saucer: "#", SaucerPadding: "-", BarStart: "[", BarEnd: "]"}),
		)
	}

	// aggregator goroutine
	var latencies []float64 // microseconds
	var mu sync.Mutex
	aggDone := make(chan struct{})
	go func() {
		for d := range latCh {
			mu.Lock()
			latencies = append(latencies, float64(d.Microseconds()))
			mu.Unlock()
		}
		close(aggDone)
	}()

	// spawn workers
	workers := r.opts.Concurrency
	perBatch := r.opts.Pipeline
	requestsTarget := r.opts.Requests
	payload := randomString(r.opts.DataSize)

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(workerId int) {
			defer wg.Done()
			for {
				// determine how many remain
				remaining := requestsTarget - int(atomic.LoadInt64(&doneCount))
				if remaining <= 0 {
					return
				}
				bsize := min(perBatch, remaining)

				startBatch := time.Now()
				if bsize == 1 && perBatch == 1 {
					if err := r.execOne(ctx, client, test, payload); err != nil && !isNilErr(err) {
						return
					}
				} else {
					if err := r.execPipeline(ctx, client, test, payload, bsize); err != nil && !isNilErr(err) {
						return
					}
				}
				elapsed := time.Since(startBatch)
				// record one latency per logical request
				perReq := elapsed / time.Duration(bsize)
				for j := 0; j < bsize; j++ {
					select {
					case latCh <- perReq:
					default:
						// drop sample on backpressure
					}
				}
				atomic.AddInt64(&doneCount, int64(bsize))
				if bar != nil { _ = bar.Add(bsize) }
			}
		}(i)
	}

	wg.Wait()
	close(latCh)
	<-aggDone
	if bar != nil { _ = bar.Finish(); fmt.Println() }

	totalElapsed := time.Since(start)

	sort.Float64s(latencies)
	p50 := percentile(latencies, 50)
	p95 := percentile(latencies, 95)
	p99 := percentile(latencies, 99)
	ops := float64(r.opts.Requests) / totalElapsed.Seconds()

	res := Result{
		TestName:        test,
		Requests:        r.opts.Requests,
		Concurrency:     r.opts.Concurrency,
		Pipeline:        r.opts.Pipeline,
		PayloadBytes:    r.opts.DataSize,
		Elapsed:         totalElapsed,
		OpsPerSec:       ops,
		ThroughputReqPS: ops, // same metric name for clarity
		P50us:           p50,
		P95us:           p95,
		P99us:           p99,
	}

	r.printSummary(res)
	return res, nil
}

func (r *Runner) execOne(ctx context.Context, client redis.UniversalClient, test, payload string) error {
	switch {
	case test == "PING":
		return client.Ping(ctx).Err()
	case test == "GET":
		_ = client.Get(ctx, r.pickKeyS()).Val(); return nil
	case test == "HGET":
		key := r.pickKeyH(); field := r.pickHashField(); _ = client.HGet(ctx, key, field).Val(); return nil
	case test == "HGETALL":
		key := r.pickKeyH(); _ = client.HGetAll(ctx, key).Val(); return nil
	case test == "LINDEX":
		key := r.pickKeyL(); idx := r.pickIndex(r.opts.ListLen); _ = client.LIndex(ctx, key, int64(idx)).Val(); return nil
	case strings.HasPrefix(test, "LRANGE_"):
		key := r.pickKeyL(); n := r.opts.RangeCount; _ = client.LRange(ctx, key, 0, int64(n-1)).Val(); return nil
	case test == "LLEN":
		key := r.pickKeyL(); _ = client.LLen(ctx, key).Val(); return nil
	case test == "SISMEMBER":
		key := r.pickKeySet(); mem := r.pickSetMember(); _, _ = client.SIsMember(ctx, key, mem).Result(); return nil
	case test == "SMEMBERS":
		key := r.pickKeySet(); _ = client.SMembers(ctx, key).Val(); return nil
	case test == "SCARD":
		key := r.pickKeySet(); _ = client.SCard(ctx, key).Val(); return nil
	case test == "ZSCORE":
		key := r.pickKeyZ(); mem := r.pickZMember(); _, _ = client.ZScore(ctx, key, mem).Result(); return nil
	case strings.HasPrefix(test, "ZRANGE_"):
		key := r.pickKeyZ(); n := r.opts.RangeCount; _ = client.ZRange(ctx, key, 0, int64(n-1)).Val(); return nil
	case test == "ZRANK":
		key := r.pickKeyZ(); mem := r.pickZMember(); _, _ = client.ZRank(ctx, key, mem).Result(); return nil
	case test == "ZCARD":
		key := r.pickKeyZ(); _ = client.ZCard(ctx, key).Val(); return nil
	case test == "GETBIT":
		key := r.pickKeyB(); bit := r.pickIndex(r.opts.BitmapSizeBits); _, _ = client.GetBit(ctx, key, int64(bit)).Result(); return nil
	case test == "BITCOUNT":
		key := r.pickKeyB(); _, _ = client.BitCount(ctx, key, &redis.BitCount{}).Result(); return nil
	case test == "PFCOUNT":
		key := r.pickKeyPF(); _, _ = client.PFCount(ctx, key).Result(); return nil
	case test == "GEOPOS":
		key := r.pickKeyG(); mem := r.pickGeoMember(); _, _ = client.GeoPos(ctx, key, mem).Result(); return nil
	case test == "GEODIST":
		key := r.pickKeyG(); m1, m2 := r.pickGeoMember(), r.pickGeoMember(); _, _ = client.GeoDist(ctx, key, m1, m2, "km").Result(); return nil
	case strings.HasPrefix(test, "XRANGE_"):
		key := r.pickKeyX(); n := r.opts.RangeCount; _ = client.XRangeN(ctx, key, "-", "+", int64(n)).Val(); return nil
	default:
		// Fallback basic write tests to keep compatibility
		switch test {
		case "SET":
			// Use configured TTL for SET operations to ensure cleanup
			key := r.pickKeyS()
			// First set the value without TTL
			if err := client.Set(ctx, key, payload, 0).Err(); err != nil {
				return err
			}
			// Then set TTL using EXPIRE command
			return client.Expire(ctx, key, r.opts.TTL).Err()
		case "GETSET":
			_, _ = client.GetSet(ctx, r.pickKeyS(), payload).Result(); return nil
		}
		return fmt.Errorf("unsupported test: %s", test)
	}
}

func (r *Runner) execPipeline(ctx context.Context, client redis.UniversalClient, test, payload string, n int) error {
	pipe := client.Pipeline()
	switch {
	case test == "PING":
		for i := 0; i < n; i++ { pipe.Ping(ctx) }
	case test == "GET":
		for i := 0; i < n; i++ { pipe.Get(ctx, r.pickKeyS()) }
	case test == "HGET":
		for i := 0; i < n; i++ { pipe.HGet(ctx, r.pickKeyH(), r.pickHashField()) }
	case test == "HGETALL":
		for i := 0; i < n; i++ { pipe.HGetAll(ctx, r.pickKeyH()) }
	case test == "LINDEX":
		for i := 0; i < n; i++ { pipe.LIndex(ctx, r.pickKeyL(), int64(r.pickIndex(r.opts.ListLen))) }
	case strings.HasPrefix(test, "LRANGE_"):
		for i := 0; i < n; i++ { pipe.LRange(ctx, r.pickKeyL(), 0, int64(r.opts.RangeCount-1)) }
	case test == "LLEN":
		for i := 0; i < n; i++ { pipe.LLen(ctx, r.pickKeyL()) }
	case test == "SISMEMBER":
		for i := 0; i < n; i++ { pipe.SIsMember(ctx, r.pickKeySet(), r.pickSetMember()) }
	case test == "SMEMBERS":
		for i := 0; i < n; i++ { pipe.SMembers(ctx, r.pickKeySet()) }
	case test == "SCARD":
		for i := 0; i < n; i++ { pipe.SCard(ctx, r.pickKeySet()) }
	case test == "ZSCORE":
		for i := 0; i < n; i++ { pipe.ZScore(ctx, r.pickKeyZ(), r.pickZMember()) }
	case strings.HasPrefix(test, "ZRANGE_"):
		for i := 0; i < n; i++ { pipe.ZRange(ctx, r.pickKeyZ(), 0, int64(r.opts.RangeCount-1)) }
	case test == "ZRANK":
		for i := 0; i < n; i++ { pipe.ZRank(ctx, r.pickKeyZ(), r.pickZMember()) }
	case test == "ZCARD":
		for i := 0; i < n; i++ { pipe.ZCard(ctx, r.pickKeyZ()) }
	case test == "GETBIT":
		for i := 0; i < n; i++ { pipe.GetBit(ctx, r.pickKeyB(), int64(r.pickIndex(r.opts.BitmapSizeBits))) }
	case test == "BITCOUNT":
		for i := 0; i < n; i++ { pipe.BitCount(ctx, r.pickKeyB(), &redis.BitCount{}) }
	case test == "PFCOUNT":
		for i := 0; i < n; i++ { pipe.PFCount(ctx, r.pickKeyPF()) }
	case test == "GEOPOS":
		for i := 0; i < n; i++ { pipe.GeoPos(ctx, r.pickKeyG(), r.pickGeoMember()) }
	case test == "GEODIST":
		for i := 0; i < n; i++ { pipe.GeoDist(ctx, r.pickKeyG(), r.pickGeoMember(), r.pickGeoMember(), "km") }
	case strings.HasPrefix(test, "XRANGE_"):
		for i := 0; i < n; i++ { pipe.XRangeN(ctx, r.pickKeyX(), "-", "+", int64(r.opts.RangeCount)) }
	default:
		switch test {
		case "SET":
			// Use configured TTL for SET operations to ensure cleanup
			// Debug: print TTL value to verify it's being set correctly
			if n == 1 {
				fmt.Printf("DEBUG: Setting TTL to %v for SET operations\n", r.opts.TTL)
			}
			// For pipeline mode, we need to use a different approach
			// Since we can't mix pipeline and individual commands, we'll execute individually
			for i := 0; i < n; i++ { 
				key := r.pickKeyS()
				// First set the value without TTL
				if err := client.Set(ctx, key, payload, 0).Err(); err != nil {
					return err
				}
				// Then set TTL using EXPIRE command
				if err := client.Expire(ctx, key, r.opts.TTL).Err(); err != nil {
					return err
				}
			}
			return nil // Return early since we're not using pipeline for this test
		case "GETSET":
			for i := 0; i < n; i++ { pipe.GetSet(ctx, r.pickKeyS(), payload) }
		default:
			return fmt.Errorf("unsupported test: %s", test)
		}
	}
	_, err := pipe.Exec(ctx)
	return err
}

func (r *Runner) printSummary(res Result) {
	fmt.Printf("\n====== %s ======\n", res.TestName)
	fmt.Printf("  %d requests completed in %.2f seconds\n", res.Requests, res.Elapsed.Seconds())
	fmt.Printf("  %d parallel clients\n", res.Concurrency)
	fmt.Printf("  %d bytes payload\n", res.PayloadBytes)
	fmt.Printf("  keep alive: 1\n")
	fmt.Printf("\n"+
		"Latency Distribution (usec):\n"+
		"  50%%\t%.2f\n"+
		"  95%%\t%.2f\n"+
		"  99%%\t%.2f\n",
		res.P50us, res.P95us, res.P99us,
	)
	fmt.Printf("\nThroughput: %.2f requests per second\n\n", res.ThroughputReqPS)
}

func percentile(sortedMicros []float64, p int) float64 {
	if len(sortedMicros) == 0 { return 0 }
	if p <= 0 { return sortedMicros[0] }
	if p >= 100 { return sortedMicros[len(sortedMicros)-1] }
	pos := (float64(p) / 100.0) * float64(len(sortedMicros)-1)
	lo := int(math.Floor(pos))
	hi := int(math.Ceil(pos))
	if lo == hi { return sortedMicros[lo] }
	weight := pos - float64(lo)
	return sortedMicros[lo]*(1-weight) + sortedMicros[hi]*weight
}

// Key helpers per type
func (r *Runner) pickKeyS() string { return r.keyS(r.pickKeyIndex()) }
func (r *Runner) pickKeyH() string { return r.keyH(r.pickKeyIndex()) }
func (r *Runner) pickKeyL() string { return r.keyL(r.pickKeyIndex()) }
func (r *Runner) pickKeySet() string { return r.keySet(r.pickKeyIndex()) }
func (r *Runner) pickKeyZ() string { return r.keyZ(r.pickKeyIndex()) }
func (r *Runner) pickKeyB() string { return r.keyB(r.pickKeyIndex()) }
func (r *Runner) pickKeyPF() string { return r.keyPF(r.pickKeyIndex()) }
func (r *Runner) pickKeyG() string { return r.keyG(r.pickKeyIndex()) }
func (r *Runner) pickKeyX() string { return r.keyX(r.pickKeyIndex()) }

func (r *Runner) keyS(i int) string  { return fmt.Sprintf("{gorb}:s:%d", i) }
func (r *Runner) keyH(i int) string  { return fmt.Sprintf("{gorb}:h:%d", i) }
func (r *Runner) keyL(i int) string  { return fmt.Sprintf("{gorb}:l:%d", i) }
func (r *Runner) keySet(i int) string { return fmt.Sprintf("{gorb}:set:%d", i) }
func (r *Runner) keyZ(i int) string  { return fmt.Sprintf("{gorb}:z:%d", i) }
func (r *Runner) keyB(i int) string  { return fmt.Sprintf("{gorb}:b:%d", i) }
func (r *Runner) keyPF(i int) string { return fmt.Sprintf("{gorb}:pf:%d", i) }
func (r *Runner) keyG(i int) string  { return fmt.Sprintf("{gorb}:g:%d", i) }
func (r *Runner) keyX(i int) string  { return fmt.Sprintf("{gorb}:x:%d", i) }

func (r *Runner) pickKeyIndex() int {
	if r.opts.Randomize {
		return rand.Intn(max(1, r.opts.Keyspace))
	}
	// simple round-robin by time
	return int(time.Now().UnixNano()) % max(1, r.opts.Keyspace)
}

func (r *Runner) pickHashField() string { return fmt.Sprintf("f%d", r.pickIndex(r.opts.HashFields)) }
func (r *Runner) pickSetMember() string { return fmt.Sprintf("m%d", r.pickIndex(r.opts.SetCardinality)) }
func (r *Runner) pickZMember() string { return fmt.Sprintf("z%d", r.pickIndex(r.opts.ZSetCardinality)) }
func (r *Runner) pickGeoMember() string { return fmt.Sprintf("g%d", r.pickIndex(r.opts.GeoMembers)) }

func (r *Runner) pickIndex(n int) int {
	if n <= 1 { return 0 }
	if r.opts.Randomize { return rand.Intn(n) }
	return int(time.Now().UnixNano()) % n
}

func randomString(n int) string {
	if n <= 0 { return "" }
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	b := make([]rune, n)
	for i := range b { b[i] = letters[rand.Intn(len(letters))] }
	return string(b)
}

func min(a, b int) int { if a < b { return a }; return b }
func max(a, b int) int { if a > b { return a }; return b }

func isNilErr(err error) bool { return err == redis.Nil }

// helpers
func containsTest(tests []string, name string) bool {
	for _, t := range tests { if strings.EqualFold(t, name) { return true } }
	return false
}

func hasAny(tests []string, prefixes ...string) bool {
	for _, t := range tests {
		upper := strings.ToUpper(strings.TrimSpace(t))
		for _, p := range prefixes { if strings.HasPrefix(upper, p) { return true } }
	}
	return false
} 