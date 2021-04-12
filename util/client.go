package util

import (
	"crypto/tls"
	"time"

	"github.com/valyala/fasthttp"
)

var (
	DefaultClient = &fasthttp.Client{
		NoDefaultUserAgentHeader:      true,
		DisablePathNormalizing:        false,
		DisableHeaderNamesNormalizing: false,
		ReadTimeout:                   5 * time.Second,
		WriteTimeout:                  5 * time.Second,
		TLSConfig: &tls.Config{
			InsecureSkipVerify: true, // per default certificates are not validated
		},
		//		Dial: func(addr string) (net.Conn, error) {
		//			return tls.Dial("tcp", addr, nil)
		//		},
		MaxConnsPerHost:           8,
		MaxIdleConnDuration:       30 * time.Second,
		MaxConnDuration:           0, // unlimited
		MaxIdemponentCallAttempts: 2,
	}
)
