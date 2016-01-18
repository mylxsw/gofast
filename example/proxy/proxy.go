package proxy

import (
	"bytes"
	"log"
	"net"
	"net/http"

	"github.com/yookoala/gofast"
)

// Proxy is the interface for a FastCGI proxy
type Proxy interface {
	ServeHTTP(w http.ResponseWriter, r *http.Request)
}

// New returns a new Proxy interface
func New(network, address string) Proxy {
	return &proxy{
		network: network,
		address: address,
	}
}

// proxy implements Proxy interface
type proxy struct {
	network string
	address string
}

// ServeHTTP implements http.Handler
func (p *proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := net.Dial(p.network, p.address)
	if err != nil {
		http.Error(w, "failed to connect to FastCGI application", http.StatusBadGateway)
		log.Printf("gofast: unable to connect to FastCGI application "+
			"(network=%#v, address=%#v, error=%#v)",
			p.network, p.address, err.Error())
		return
	}

	c := gofast.NewClient(conn)
	req := c.NewRequest()

	// some required cgi parameters
	req.Params["REQUEST_METHOD"] = r.Method
	req.Params["SERVER_PROTOCOL"] = r.Proto

	/*
		// FIXME: add these parameter automatically
		// from net/cgi Handler.ServeHTTP
		// should add later
		"SERVER_SOFTWARE=go",
		"SERVER_NAME=" + req.Host,
		"SERVER_PROTOCOL=HTTP/1.1",
		"HTTP_HOST=" + req.Host,
		"GATEWAY_INTERFACE=CGI/1.1",
		"REQUEST_METHOD=" + req.Method,
		"QUERY_STRING=" + req.URL.RawQuery,
		"REQUEST_URI=" + req.URL.RequestURI(),
		"PATH_INFO=" + pathInfo,
		"SCRIPT_NAME=" + root,
		"SCRIPT_FILENAME=" + h.Path,
		"SERVER_PORT=" + port,
	*/

	// some input for req
	req.Params["hello"] = "world"
	req.Params["foo"] = "bar"
	req.Content = []byte("foo bar") // FIXME: somehow need content, find out why

	// a buffer to read
	wResp := new(bytes.Buffer)
	wErr := new(bytes.Buffer)

	// handl the result
	err = c.Handle(wResp, wErr, req)
	if err != nil {
		http.Error(w, "failed to process request", http.StatusInternalServerError)
		log.Printf("gofast: unable to process request "+
			"(network=%#v, address=%#v, error=%#v)",
			p.network, p.address, err.Error())
		return
	} else if wErr.Len() != 0 {
		log.Printf("gofast: error stream from FastCGI application "+
			"(network=%#v, address=%#v, error=%#v)",
			p.network, p.address, wErr.String())
	}
	w.Write(wResp.Bytes())
}
