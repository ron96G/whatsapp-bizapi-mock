package util

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
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
		MaxConnsPerHost:           8,
		MaxIdleConnDuration:       30 * time.Second,
		MaxConnDuration:           0, // unlimited
		MaxIdemponentCallAttempts: 2,
	}
)

func NewClient(rootCa []byte) error {

	if rootCa != nil {
		caCertPool := x509.NewCertPool()

		if !caCertPool.AppendCertsFromPEM(rootCa) {
			return fmt.Errorf("unable to parse provided certificate")
		}

		DefaultClient.TLSConfig = &tls.Config{
			RootCAs:            caCertPool,
			InsecureSkipVerify: false,
		}
	}

	return nil
}
