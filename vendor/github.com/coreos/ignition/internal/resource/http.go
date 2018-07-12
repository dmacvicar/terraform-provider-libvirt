// Copyright 2016 CoreOS, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package resource

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"io"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/coreos/ignition/internal/config/types"
	"github.com/coreos/ignition/internal/log"
	"github.com/coreos/ignition/internal/util"
	"github.com/coreos/ignition/internal/version"

	"github.com/vincent-petithory/dataurl"
)

const (
	initialBackoff = 100 * time.Millisecond
	maxBackoff     = 5 * time.Second

	defaultHttpResponseHeaderTimeout = 10
	defaultHttpTotalTimeout          = 0
)

var (
	ErrTimeout         = errors.New("unable to fetch resource in time")
	ErrPEMDecodeFailed = errors.New("unable to decode PEM block")
)

// HttpClient is a simple wrapper around the Go HTTP client that standardizes
// the process and logging of fetching payloads.
type HttpClient struct {
	client  *http.Client
	logger  *log.Logger
	timeout time.Duration

	transport *http.Transport
	cas       map[types.CaReference][]byte
}

func (f *Fetcher) UpdateHttpTimeoutsAndCAs(timeouts types.Timeouts, cas []types.CaReference) error {
	if f.client == nil {
		f.newHttpClient()
	}

	// Update timeouts
	responseHeader := defaultHttpResponseHeaderTimeout
	total := defaultHttpTotalTimeout
	if timeouts.HTTPResponseHeaders != nil {
		responseHeader = *timeouts.HTTPResponseHeaders
	}
	if timeouts.HTTPTotal != nil {
		total = *timeouts.HTTPTotal
	}

	f.client.client.Timeout = time.Duration(total) * time.Second
	f.client.timeout = f.client.client.Timeout

	f.client.transport.ResponseHeaderTimeout = time.Duration(responseHeader) * time.Second
	f.client.client.Transport = f.client.transport

	// Update CAs
	if len(cas) == 0 {
		return nil
	}

	pool, err := x509.SystemCertPool()
	if err != nil {
		f.Logger.Err("Unable to read system certificate pool: %s", err)
		return err
	}

	for _, ca := range cas {
		cablob, err := f.getCABlob(ca)
		if err != nil {
			return err
		}
		block, _ := pem.Decode(cablob)
		if block == nil {
			f.Logger.Err("Unable to decode CA (%s)", ca.Source)
			return ErrPEMDecodeFailed
		}

		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			f.Logger.Err("Unable to parse CA (%s): %s", ca.Source, err)
			return err
		}

		f.Logger.Info("Adding %q to list of CAs", cert.Subject.CommonName)
		pool.AddCert(cert)
	}

	f.client.transport.TLSClientConfig = &tls.Config{RootCAs: pool}
	f.client.client.Transport = f.client.transport
	return nil
}

func (f *Fetcher) getCABlob(ca types.CaReference) ([]byte, error) {
	if blob, ok := f.client.cas[ca]; ok {
		return blob, nil
	}
	u, err := url.Parse(ca.Source)
	if err != nil {
		f.Logger.Crit("Unable to parse CA URL: %s", err)
		return nil, err
	}
	hasher, err := util.GetHasher(ca.Verification)
	if err != nil {
		f.Logger.Crit("Unable to get hasher: %s", err)
		return nil, err
	}

	var expectedSum []byte
	if hasher != nil {
		// explicitly ignoring the error here because the config should already
		// be validated by this point
		_, expectedSumString, _ := util.HashParts(ca.Verification)
		expectedSum, err = hex.DecodeString(expectedSumString)
		if err != nil {
			f.Logger.Crit("Error parsing verification string %q: %v", expectedSumString, err)
			return nil, err
		}
	}

	cablob, err := f.FetchToBuffer(*u, FetchOptions{
		Hash:        hasher,
		ExpectedSum: expectedSum,
	})
	if err != nil {
		f.Logger.Err("Unable to fetch CA (%s): %s", u, err)
		return nil, err
	}
	f.client.cas[ca] = cablob
	return cablob, nil

}

// RewriteCAsWithDataUrls will modify the passed in slice of CA references to
// contain the actual CA file via a dataurl in their source field.
func (f *Fetcher) RewriteCAsWithDataUrls(cas []types.CaReference) error {
	for i, ca := range cas {
		blob, err := f.getCABlob(ca)
		if err != nil {
			return err
		}

		cas[i].Source = dataurl.EncodeBytes(blob)
	}
	return nil
}

func (f *Fetcher) newHttpClient() {
	transport := &http.Transport{
		ResponseHeaderTimeout: time.Duration(defaultHttpResponseHeaderTimeout) * time.Second,
		Dial: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			Resolver: &net.Resolver{
				PreferGo: true,
			},
		}).Dial,
		TLSHandshakeTimeout: 10 * time.Second,
	}
	f.client = &HttpClient{
		client: &http.Client{
			Transport: transport,
		},
		logger:    f.Logger,
		timeout:   time.Duration(defaultHttpTotalTimeout) * time.Second,
		transport: transport,
		cas:       make(map[types.CaReference][]byte),
	}
}

// getReaderWithHeader performs an HTTP GET on the provided URL with the provided request header
// and returns the response body Reader, HTTP status code, and error (if any). By
// default, User-Agent is added to the header but this can be overridden.
func (c HttpClient) getReaderWithHeader(ctx context.Context, url string, header http.Header) (io.ReadCloser, int, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, 0, err
	}

	req.Header.Set("User-Agent", "Ignition/"+version.Raw)

	for key, values := range header {
		req.Header.Del(key)
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	if c.timeout != 0 {
		ctxTo, cancel := context.WithTimeout(ctx, c.timeout)
		ctx = ctxTo
		defer cancel()
	}

	duration := initialBackoff
	for attempt := 1; ; attempt++ {
		c.logger.Info("GET %s: attempt #%d", url, attempt)
		resp, err := c.client.Do(req.WithContext(ctx))

		if err == nil {
			c.logger.Info("GET result: %s", http.StatusText(resp.StatusCode))
			if resp.StatusCode < 500 {
				return resp.Body, resp.StatusCode, nil
			}
			resp.Body.Close()
		} else {
			c.logger.Info("GET error: %v", err)
		}

		duration = duration * 2
		if duration > maxBackoff {
			duration = maxBackoff
		}

		// Wait before next attempt or exit if we timeout while waiting
		select {
		case <-time.After(duration):
		case <-ctx.Done():
			return nil, 0, ErrTimeout
		}
	}
}
