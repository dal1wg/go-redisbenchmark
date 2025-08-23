/*
 * go-redisbenchmark - A high performance Redis benchmark tool written in Go.
 * Supports Redis 6.0+ features including ACL and TLS.
 *
 * Copyright (c) 2025, dal1wg <15072697283@139.com>
 * All rights reserved.
 *
 * This source code is licensed under the BSD 3-Clause License.
 */

package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
	benchmark "goRedisBenchmark/pkg"
)

func main() {
	var (
		modeStr = flag.String("mode", string(benchmark.ModeStandalone), "mode: standalone|cluster|sentinel")

		// Connection
		uri = flag.String("uri", "", "Redis URI (e.g., redis://username:password@host:port/db?tls=true)")

		// common/standalone
		host     = flag.String("h", "127.0.0.1", "Server hostname")
		port     = flag.Int("p", 6379, "Server port")
		addr     = flag.String("addr", "", "Server address host:port (overrides -h and -p)")
		password = flag.String("a", "", "Password to use when connecting to the server")
		db       = flag.Int("db", 0, "Database number")

		// ACL support (Redis 6.0+)
		username    = flag.String("user", "", "ACL username for Redis 6.0+")
		aclPassword = flag.String("pass", "", "ACL password for Redis 6.0+ (separate from legacy password)")

		// TLS support (Redis 6.0+)
		useTLS      = flag.Bool("tls", false, "Enable TLS connection")
		tlsCertFile = flag.String("tls-cert", "", "Client certificate file path")
		tlsKeyFile  = flag.String("tls-key", "", "Client private key file path")
		tlsCAFile   = flag.String("tls-ca", "", "CA certificate file path")
		tlsInsecure = flag.Bool("tls-insecure", false, "Skip TLS certificate verification")

		// sentinel
		sentinelAddrsStr = flag.String("sentinel-addrs", "", "Sentinel nodes, comma-separated host:port list")
		sentinelMaster   = flag.String("sentinel-master", "", "Sentinel master name")

		// workload
		numRequests = flag.Int("n", 100000, "Total number of requests")
		clients     = flag.Int("c", 50, "Number of parallel connections")
		pipeline    = flag.Int("P", 1, "Pipeline requests per round-trip")
		dataSize    = flag.Int("d", 3, "Data size of SET/GET payload in bytes")
		testsStr    = flag.String("t", "PING,GET,HGET,LINDEX,SISMEMBER,ZSCORE,GETBIT,PFCOUNT,GEOPOS,XRANGE_", "Tests to run, comma separated. Use suffix _ for range types: LRANGE_, ZRANGE_, XRANGE_.")
		keyspace    = flag.Int("k", 1000, "Keyspace size for per-type keys")
		randomize   = flag.Bool("r", true, "Randomize key selection")
		quiet       = flag.Bool("q", false, "Quiet mode: only progress and summary")

		// dataset sizes
		hashFields = flag.Int("hash-fields", 10, "Fields per hash")
		listLen    = flag.Int("list-len", 10, "Elements per list")
		setCard    = flag.Int("set-card", 10, "Members per set")
		zsetCard   = flag.Int("zset-card", 10, "Members per zset")
		bitmapBits = flag.Int("bitmap-bits", 1024, "Bitmap bits length")
		geoMembers = flag.Int("geo-members", 10, "Members per geo set")
		streamLen  = flag.Int("stream-len", 10, "Entries per stream")
		rangeCount = flag.Int("range-count", 10, "Count for LRANGE/ZRANGE/XRANGE reads")

		// shortcuts
		readAll = flag.Bool("read-all", false, "Run a suite of read commands across all data types")

		// data lifecycle
		ttl = flag.Duration("ttl", time.Hour, "TTL for benchmark data (e.g., 1h, 30m, 2h)")
	)

	flag.Parse()

	// Show help if requested
	if len(os.Args) == 1 || (len(os.Args) > 1 && (os.Args[1] == "-h" || os.Args[1] == "--help")) {
		fmt.Println("go-redisbenchmark - A high performance Redis benchmark tool")
		fmt.Println("\nUsage:")
		fmt.Println("  ./go-redisbenchmark [options]")
		fmt.Println("\nACL Authentication:")
		fmt.Println("  --user <username>     ACL username for Redis 6.0+")
		fmt.Println("  --pass <password>     ACL password for Redis 6.0+")
		fmt.Println("\nTLS Options:")
		fmt.Println("  --tls                 Enable TLS connection")
		fmt.Println("  --tls-cert <file>     Client certificate file")
		fmt.Println("  --tls-key <file>      Client private key file")
		fmt.Println("  --tls-ca <file>       CA certificate file")
		fmt.Println("\nRedis URI:")
		fmt.Println("  --uri <uri>           Redis connection URI")
		fmt.Println("                         Example: redis://user:pass@host:port/db")
		fmt.Println("\nData Lifecycle:")
		fmt.Println("  --ttl <duration>      TTL for benchmark data (default: 1m)")
		fmt.Println("                         Examples: 1h, 30m, 2h, 3600s")
		fmt.Println("\nRun with --help for full options list")
		os.Exit(0)
	}

	// context with cancel on Ctrl+C
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// derive addresses
	finalAddr := *addr
	if finalAddr == "" {
		finalAddr = fmt.Sprintf("%s:%d", *host, *port)
	}

	var sentinelAddrs []string
	if strings.TrimSpace(*sentinelAddrsStr) != "" {
		sentinelAddrs = splitAndTrim(*sentinelAddrsStr)
	}

	tests := splitAndTrim(*testsStr)
	if *readAll {
		tests = []string{
			"GET",
			"HGET", "HGETALL",
			"LINDEX", "LRANGE_", "LLEN",
			"SISMEMBER", "SMEMBERS", "SCARD",
			"ZSCORE", "ZRANGE_", "ZRANK", "ZCARD",
			"GETBIT", "BITCOUNT",
			"PFCOUNT",
			"GEOPOS", "GEODIST",
			"XRANGE_",
		}
	}
	if len(tests) == 0 {
		fmt.Fprintln(os.Stderr, "no tests specified")
		os.Exit(2)
	}

	opts := benchmark.Options{
		Mode:        benchmark.Mode(strings.ToLower(strings.TrimSpace(*modeStr))),
		URI:         *uri,
		Addr:        finalAddr,
		DB:          *db,
		Password:    *password,
		Username:    *username,
		ACLPassword: *aclPassword,
		UseTLS:      *useTLS,
		TLSCertFile: *tlsCertFile,
		TLSKeyFile:  *tlsKeyFile,
		TLSCAFile:   *tlsCAFile,
		TLSInsecure: *tlsInsecure,

		SentinelAddrs:  sentinelAddrs,
		SentinelMaster: *sentinelMaster,
		Requests:       *numRequests,
		Concurrency:    *clients,
		Pipeline:       *pipeline,
		DataSize:       *dataSize,
		Tests:          tests,
		Keyspace:       *keyspace,
		Randomize:      *randomize,
		Quiet:          *quiet,

		HashFields:      *hashFields,
		ListLen:         *listLen,
		SetCardinality:  *setCard,
		ZSetCardinality: *zsetCard,
		BitmapSizeBits:  *bitmapBits,
		GeoMembers:      *geoMembers,
		StreamLen:       *streamLen,
		RangeCount:      *rangeCount,
		TTL:             *ttl,
	}

	r, err := benchmark.NewRunner(opts)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	start := time.Now()
	results, err := r.Run(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Benchmark failed: %v\n", err)
		if strings.Contains(err.Error(), "authentication") || strings.Contains(err.Error(), "password") {
			fmt.Fprintln(os.Stderr, "\nAuthentication troubleshooting:")
			fmt.Fprintln(os.Stderr, "1. Check if the username and password are correct")
			fmt.Fprintln(os.Stderr, "2. Verify Redis ACL configuration")
			fmt.Fprintln(os.Stderr, "3. Try using --user and --pass flags instead of --uri")
			fmt.Fprintln(os.Stderr, "4. Check Redis server logs for authentication errors")
		}
		os.Exit(1)
	}
	_ = results
	fmt.Printf("Total elapsed: %.2fs\n", time.Since(start).Seconds())
}

func splitAndTrim(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
