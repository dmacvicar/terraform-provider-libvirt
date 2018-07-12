// Copyright 2018 CoreOS, Inc.
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

package security

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"

	"github.com/coreos/ignition/tests/register"
	"github.com/coreos/ignition/tests/types"
)

func init() {
	cer, err := tls.X509KeyPair(publicKey, privateKey)
	if err != nil {
		panic(fmt.Sprintf("error loading x509 keypair: %v", err))
	}
	config := &tls.Config{Certificates: []tls.Certificate{cer}}
	customCAServer.TLS = config
	customCAServer.Config.ErrorLog = log.New(ioutil.Discard, "", 0)
	customCAServer.StartTLS()

	register.Register(register.NegativeTest, AppendConfigCustomCert())
	register.Register(register.NegativeTest, FetchFileCustomCert())
}

var (
	// generated via:
	// openssl ecparam -genkey -name secp384r1 -out server.key
	privateKey = []byte(`-----BEGIN EC PARAMETERS-----
BgUrgQQAIg==
-----END EC PARAMETERS-----
-----BEGIN EC PRIVATE KEY-----
MIGkAgEBBDB6yW6RIYfTXdYVuPY0V0L6EtZ6vZD86vgbsw52Y3/U5nZ2JE++JrKu
tt2Xt/NMzG6gBwYFK4EEACKhZANiAAQDEhfHEulYKlANw9eR5l455gwzAIQuraa0
49RhvM7PPywaiD8DobteQmE8wn7cJSzOYw6GLvrL4Q1BO5EFUXknkW50t8lfnUeH
veCNsqvm82F1NVevVoExAUhDYmMREa4=
-----END EC PRIVATE KEY-----`)

	// generated via:
	// openssl req -new -x509 -sha256 -key server.key -out server.crt -days 3650
	publicKey = []byte(`-----BEGIN CERTIFICATE-----
MIICzTCCAlKgAwIBAgIJALTP0pfNBMzGMAoGCCqGSM49BAMCMIGZMQswCQYDVQQG
EwJVUzETMBEGA1UECAwKQ2FsaWZvcm5pYTEWMBQGA1UEBwwNU2FuIEZyYW5jaXNj
bzETMBEGA1UECgwKQ29yZU9TIEluYzEUMBIGA1UECwwLRW5naW5lZXJpbmcxEzAR
BgNVBAMMCmNvcmVvcy5jb20xHTAbBgkqhkiG9w0BCQEWDm9lbUBjb3Jlb3MuY29t
MB4XDTE4MDEyNTAwMDczOVoXDTI4MDEyMzAwMDczOVowgZkxCzAJBgNVBAYTAlVT
MRMwEQYDVQQIDApDYWxpZm9ybmlhMRYwFAYDVQQHDA1TYW4gRnJhbmNpc2NvMRMw
EQYDVQQKDApDb3JlT1MgSW5jMRQwEgYDVQQLDAtFbmdpbmVlcmluZzETMBEGA1UE
AwwKY29yZW9zLmNvbTEdMBsGCSqGSIb3DQEJARYOb2VtQGNvcmVvcy5jb20wdjAQ
BgcqhkjOPQIBBgUrgQQAIgNiAAQDEhfHEulYKlANw9eR5l455gwzAIQuraa049Rh
vM7PPywaiD8DobteQmE8wn7cJSzOYw6GLvrL4Q1BO5EFUXknkW50t8lfnUeHveCN
sqvm82F1NVevVoExAUhDYmMREa6jZDBiMA8GA1UdEQQIMAaHBH8AAAEwHQYDVR0O
BBYEFEbFy0SPiF1YXt+9T3Jig2rNmBtpMB8GA1UdIwQYMBaAFEbFy0SPiF1YXt+9
T3Jig2rNmBtpMA8GA1UdEwEB/wQFMAMBAf8wCgYIKoZIzj0EAwIDaQAwZgIxAOul
t3MhI02IONjTDusl2YuCxMgpy2uy0MPkEGUHnUOsxmPSG0gEBCNHyeKVeTaPUwIx
AKbyaAqbChEy9CvDgyv6qxTYU+eeBImLKS3PH2uW5etc/69V/sDojqpH3hEffsOt
9g==
-----END CERTIFICATE-----`)

	customCAServerFile = []byte(`{
			"ignition": { "version": "2.1.0" },
			"storage": {
				"files": [{
					"filesystem": "root",
					"path": "/foo/bar",
					"contents": { "source": "data:,example%20file%0A" }
				}]
			}
		}`)

	customCAServer = httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(customCAServerFile)
	}))
)

func AppendConfigCustomCert() types.Test {
	name := "Append config with custom tls cert"
	in := types.GetBaseDisk()
	out := types.GetBaseDisk()
	config := fmt.Sprintf(`{
		"ignition": {
			"version": "2.2.0",
			"config": {
			  "append": [{
				"source": %q
			  }]
			},
			"timeouts": {
				"httpTotal": 5
			}
		}
	}`, customCAServer.URL)

	return types.Test{
		Name:   name,
		In:     in,
		Out:    out,
		Config: config,
	}
}

func FetchFileCustomCert() types.Test {
	name := "Fetch file with custom tls cert"
	in := types.GetBaseDisk()
	out := types.GetBaseDisk()
	config := fmt.Sprintf(`{
		"ignition": {
			"version": "2.2.0",
			"timeouts": {
				"httpTotal": 5
			}
		},
		"storage": {
			"files": [{
				"filesystem": "root",
				"path": "/foo/bar",
				"contents": {
					"source": %q
				}
			}]
		}
	}`, customCAServer.URL)

	return types.Test{
		Name:   name,
		In:     in,
		Out:    out,
		Config: config,
	}
}
