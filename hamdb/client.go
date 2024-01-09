package hamdb

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	backoff "github.com/cenkalti/backoff/v4"
	ttlcache "github.com/jellydator/ttlcache/v3"

	"github.com/charmbracelet/log"
)

var (
	DefaultTimeout    = 5 * time.Second
	DefaultAppName    = "go-hamdb"
	DefaultBaseURL    = "https://api.hamdb.org/v1"
	DefaultCacheTTL   = 24 * time.Hour
	DefaultCacheNXTTL = 5 * time.Minute
	Logger            = log.Default()
)

// backoffStrategy is our standard retry strategy for s3 operations. The
// slightly unusual executed function lets us build on the default settings
// provided by the backoff package.
var backoffStrategy = func() *backoff.ExponentialBackOff {
	b := backoff.NewExponentialBackOff()

	b.InitialInterval = 1 * time.Second
	b.MaxInterval = 15 * time.Second

	return b
}()

type Client struct {
	baseURL    string
	httpClient *http.Client
	appName    string

	log        *log.Logger
	cache      *ttlcache.Cache[string, Response]
	nxCacheTTL time.Duration
}

type OptionFunc func(*Client)

func WithClient(c *http.Client) OptionFunc {
	// WithClient defines the http client used to make requests to the hamdb api.
	return func(h *Client) {
		h.httpClient = c
	}
}

func WithBaseURL(url string) OptionFunc {
	// WithBaseURL defines the base url used to make requests to the hamdb api.
	return func(h *Client) {
		h.baseURL = url
	}
}

func WithLogger(l *log.Logger) OptionFunc {
	// WithLogger defines the logger used by the hamdb client.
	return func(h *Client) {
		h.log = l
	}
}

func WithCacheTTL(ttl time.Duration) OptionFunc {
	// WithCacheTTL defines the ttl of the cache used by the hamdb client.
	return func(h *Client) {
		h.cache = ttlcache.New[string, Response](
			ttlcache.WithTTL[string, Response](ttl),
		)
	}
}

func WithNXCacheTTL(ttl time.Duration) OptionFunc {
	// WithNXCacheTTL defines the ttl of the cache used by the hamdb client for
	// callsigns that are not found.
	return func(h *Client) {
		h.nxCacheTTL = ttl
	}
}

func New(opts ...OptionFunc) *Client {
	// New returns a new hamdb client.
	// It also starts the cache cleaner goroutine. It's probably a good idea to
	// the Stop function to halt the cleaner if the client is no longer needed.
	c := &Client{
		appName: DefaultAppName,
		baseURL: DefaultBaseURL,
		httpClient: &http.Client{
			Timeout: DefaultTimeout,
		},
		log: Logger.With("service", "hamdb"),
		cache: ttlcache.New[string, Response](
			ttlcache.WithTTL[string, Response](DefaultCacheTTL),
		),
		nxCacheTTL: DefaultCacheNXTTL,
	}

	c.log.SetLevel(log.DebugLevel)

	for _, opt := range opts {
		opt(c)
	}

	c.log = c.log.With("pkg", "hamdb")

	go c.cache.Start()

	return c
}

func (c *Client) ClearCache() {
	c.cache.DeleteAll()
}

func (c *Client) Stop() {
	c.cache.Stop()
}

func (c *Client) lookup(ctx context.Context, callsign string) (Response, error) {
	var resp Response
	callsign = strings.TrimSpace(callsign)
	url := fmt.Sprintf("%s/%s/json/%s", c.baseURL, callsign, c.appName)

	l := c.log.With("callsign", callsign, "url", url)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return resp, err
	}
	req.Header.Set("User-Agent", c.appName)

	l.Debug("starting lookup")
	r, err := c.httpClient.Do(req)
	if err != nil {
		return resp, err
	}
	defer r.Body.Close()

	if r.StatusCode != http.StatusOK {
		l.Error("request failed", "status", r.StatusCode)

		return resp, fmt.Errorf("request failed with status %d", r.StatusCode)
	}

	if err = json.NewDecoder(r.Body).Decode(&resp); err != nil {
		l.Error("json decode failed", "error", err)
		return resp, err
	}

	return resp, resp.Status()
}

func (c *Client) Lookup(ctx context.Context, callsign string) (Callsign, error) {
	var resp Response

	l := c.log.With("callsign", callsign)

	if c.cache.Has(callsign) {
		hit := c.cache.Get(callsign).Value()
		l.Debug("cache hit", "nx", hit.Status())
		return hit.HamDB.Callsign, hit.Status()
	}

	l.Debug("cache miss")

	op := func() error {
		r, err := c.lookup(ctx, callsign)
		resp = r
		if err != nil {
			if err == ErrNotFound {
				l.Debug("callsign not found")
				return &backoff.PermanentError{Err: err}
			}

			l.Warn("retrying", "error", err)
		}
		return err
	}
	err := backoff.Retry(op, backoffStrategy)
	if err == nil {
		l.Debug("cache set")
		c.cache.Set(callsign, resp, ttlcache.DefaultTTL)
	} else {
		l.Debug("nxcache set")
		c.cache.Set(callsign, resp, c.nxCacheTTL)
	}

	return resp.HamDB.Callsign, err
}

var defaultClient *Client

func Lookup(ctx context.Context, callsign string) (Callsign, error) {
	if defaultClient == nil {
		defaultClient = New()
	}
	return defaultClient.Lookup(ctx, callsign)
}
