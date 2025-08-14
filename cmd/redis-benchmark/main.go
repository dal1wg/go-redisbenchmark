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

	"go-redisbenchmark/internal/benchmark"
)

func main() {
	var (
		modeStr          = flag.String("mode", string(benchmark.ModeStandalone), "mode: standalone|cluster|sentinel")

		// common/standalone
		host             = flag.String("h", "127.0.0.1", "Server hostname")
		port             = flag.Int("p", 6379, "Server port")
		addr             = flag.String("addr", "", "Server address host:port (overrides -h and -p)")
		password         = flag.String("a", "", "Password to use when connecting to the server")
		db               = flag.Int("db", 0, "Database number")

		// cluster
		clusterAddrsStr = flag.String("cluster-addrs", "", "Cluster startup nodes, comma-separated host:port list")

		// sentinel
		sentinelAddrsStr = flag.String("sentinel-addrs", "", "Sentinel nodes, comma-separated host:port list")
		sentinelMaster   = flag.String("sentinel-master", "", "Sentinel master name")

		// workload
		numRequests      = flag.Int("n", 100000, "Total number of requests")
		clients          = flag.Int("c", 50, "Number of parallel connections")
		pipeline         = flag.Int("P", 1, "Pipeline requests per round-trip")
		dataSize         = flag.Int("d", 3, "Data size of SET/GET payload in bytes")
		testsStr         = flag.String("t", "PING,GET,HGET,LINDEX,SISMEMBER,ZSCORE,GETBIT,PFCOUNT,GEOPOS,XRANGE_", "Tests to run, comma separated. Use suffix _ for range types: LRANGE_, ZRANGE_, XRANGE_.")
		keyspace         = flag.Int("k", 1000, "Keyspace size for per-type keys")
		randomize        = flag.Bool("r", true, "Randomize key selection")
		quiet            = flag.Bool("q", false, "Quiet mode: only progress and summary")

		// dataset sizes
		hashFields       = flag.Int("hash-fields", 10, "Fields per hash")
		listLen          = flag.Int("list-len", 10, "Elements per list")
		setCard          = flag.Int("set-card", 10, "Members per set")
		zsetCard         = flag.Int("zset-card", 10, "Members per zset")
		bitmapBits       = flag.Int("bitmap-bits", 1024, "Bitmap bits length")
		geoMembers       = flag.Int("geo-members", 10, "Members per geo set")
		streamLen        = flag.Int("stream-len", 10, "Entries per stream")
		rangeCount       = flag.Int("range-count", 10, "Count for LRANGE/ZRANGE/XRANGE reads")

		// shortcuts
		readAll          = flag.Bool("read-all", false, "Run a suite of read commands across all data types")
	)

	flag.Parse()

	// context with cancel on Ctrl+C
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// derive addresses
	finalAddr := *addr
	if finalAddr == "" {
		finalAddr = fmt.Sprintf("%s:%d", *host, *port)
	}

	var clusterAddrs []string
	if strings.TrimSpace(*clusterAddrsStr) != "" {
		clusterAddrs = splitAndTrim(*clusterAddrsStr)
	}

	var sentinelAddrs []string
	if strings.TrimSpace(*sentinelAddrsStr) != "" {
		sentinelAddrs = splitAndTrim(*sentinelAddrsStr)
	}

	tests := splitAndTrim(*testsStr)
	if *readAll {
		tests = []string{
			"GET",
			"HGET","HGETALL",
			"LINDEX","LRANGE_","LLEN",
			"SISMEMBER","SMEMBERS","SCARD",
			"ZSCORE","ZRANGE_","ZRANK","ZCARD",
			"GETBIT","BITCOUNT",
			"PFCOUNT",
			"GEOPOS","GEODIST",
			"XRANGE_",
		}
	}
	if len(tests) == 0 {
		fmt.Fprintln(os.Stderr, "no tests specified")
		os.Exit(2)
	}

	opts := benchmark.Options{
		Mode:           benchmark.Mode(strings.ToLower(strings.TrimSpace(*modeStr))),
		Addr:           finalAddr,
		DB:             *db,
		Password:       *password,
		ClusterAddrs:   clusterAddrs,
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
	}

	r, err := benchmark.NewRunner(opts)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	start := time.Now()
	results, err := r.Run(ctx)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
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