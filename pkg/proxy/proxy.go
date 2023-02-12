package proxy

import (
	"net/http"
	"net/http/httptrace"
	"net/http/httputil"

	"github.com/dddpaul/http-over-socks-proxy/pkg/logger"
	"github.com/dddpaul/http-over-socks-proxy/pkg/transport"
	log "github.com/sirupsen/logrus"
)

type Proxy struct {
	proxy     *httputil.ReverseProxy
	port      string
	transport http.RoundTripper
	trace     bool
}

type ProxyOption func(p *Proxy)

func WithPort(port string) ProxyOption {
	return func(p *Proxy) {
		p.port = port
	}
}

func WithSocks(socks string) ProxyOption {
	return func(p *Proxy) {
		p.transport = transport.NewSocksTransport(socks)
	}
}

func WithTrace(enabled bool) ProxyOption {
	return func(p *Proxy) {
		p.trace = enabled
	}
}

func New(opts ...ProxyOption) *Proxy {
	p := &Proxy{}

	for _, opt := range opts {
		opt(p)
	}

	director := func(req *http.Request) {
		ctx := req.Context()
		logger.Log(ctx, nil).WithFields(log.Fields{
			"request":    req.RequestURI,
			"method":     req.Method,
			"remote":     req.RemoteAddr,
			"user-agent": req.UserAgent(),
			"referer":    req.Referer(),
		}).Debugf("request")
		if p.trace {
			r := req.WithContext(httptrace.WithClientTrace(ctx, transport.NewTrace(ctx)))
			*req = *r
		}
	}

	p.proxy = &httputil.ReverseProxy{
		Transport: p.transport,
		Director:  director,
	}

	return p
}

func (p *Proxy) Start() {
	log.Infof("Start HTTP proxy on port %s", p.port)
	if err := http.ListenAndServe(p.port, p.proxy); err != nil {
		panic(err)
	}
}
