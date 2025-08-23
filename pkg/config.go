package pkg

import "time"

// Mode represents Redis deployment mode.
type Mode string

const (
	ModeStandalone Mode = "standalone"
	ModeCluster    Mode = "cluster"
	ModeSentinel   Mode = "sentinel"
)

// Options configures a benchmark run.
type Options struct {
	Mode Mode

	// Connection
	URI string // Redis URI (e.g., redis://username:password@host:port/db?tls=true)

	// Standalone
	Addr string // host:port
	DB   int

	// Shared
	Password string

	// ACL support (Redis 6.0+)
	Username    string // ACL username
	ACLPassword string // ACL password (separate from legacy password)

	// TLS support (Redis 6.0+)
	UseTLS      bool   // Enable TLS connection
	TLSCertFile string // Client certificate file path
	TLSKeyFile  string // Client private key file path
	TLSCAFile   string // CA certificate file path
	TLSInsecure bool   // Skip TLS certificate verification

	// Sentinel
	SentinelAddrs  []string // host:port list
	SentinelMaster string

	// Workload
	Requests    int      // total number of requests per test
	Concurrency int      // parallel workers
	Pipeline    int      // number of pipelined commands per round-trip
	DataSize    int      // payload size for SET/GET
	Tests       []string // e.g., ["PING","SET","GET","HGET",...]
	Keyspace    int      // number of distinct keys per type
	Randomize   bool     // randomize keys each request
	Quiet       bool     // suppress per-interval logs (still shows progress)

	// Dataset sizes for complex types
	HashFields      int // fields per hash
	ListLen         int // elements per list
	SetCardinality  int // members per set
	ZSetCardinality int // members per zset
	BitmapSizeBits  int // bitmap length in bits
	GeoMembers      int // members per geo set
	StreamLen       int // entries per stream
	RangeCount      int // count for range reads (LRANGE/ZRANGE/XRANGE)

	// Data lifecycle management
	TTL time.Duration // TTL for benchmark data (default: 1 hour)
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
	P50us float64
	P95us float64
	P99us float64
}

// Runner executes benchmarks based on Options.
type Runner struct {
	opts Options
}
