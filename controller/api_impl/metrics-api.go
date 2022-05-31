/*
	Copyright NetFoundry, Inc.

	Licensed under the Apache License, Version 2.0 (the "License");
	you may not use this file except in compliance with the License.
	You may obtain a copy of the License at

	https://www.apache.org/licenses/LICENSE-2.0

	Unless required by applicable law or agreed to in writing, software
	distributed under the License is distributed on an "AS IS" BASIS,
	WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
	See the License for the specific language governing permissions and
	limitations under the License.
*/

package api_impl

import (
	"bytes"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"github.com/michaelquigley/pfxlog"
	"github.com/openziti/channel"
	"github.com/openziti/fabric/controller/network"
	"github.com/openziti/fabric/controller/xmgmt"
	"github.com/openziti/foundation/identity/identity"
	"github.com/openziti/xweb"
	"net/http"
	"regexp"
	"strings"
)

var _ xweb.WebHandlerFactory = &MetricsApiFactory{}

type MetricsApiFactory struct {
	network *network.Network
	nodeId  identity.Identity
	xmgmts  []xmgmt.Xmgmt
}

func (factory *MetricsApiFactory) Validate(_ *xweb.Config) error {
	return nil
}

func NewMetricsApiFactory(nodeId identity.Identity, network *network.Network, xmgmts []xmgmt.Xmgmt) *MetricsApiFactory {
	return &MetricsApiFactory{
		network: network,
		nodeId:  nodeId,
		xmgmts:  xmgmts,
	}
}

func (factory *MetricsApiFactory) Binding() string {
	return MetricApiBinding
}

func (factory *MetricsApiFactory) New(_ *xweb.WebListener, options map[interface{}]interface{}) (xweb.WebHandler, error) {

	metricsApiHandler, err := NewMetricsApiHandler(factory.network, options)

	if err != nil {
		return nil, err
	}

	return metricsApiHandler, nil
}

func NewMetricsApiHandler(n *network.Network, options map[interface{}]interface{}) (*MetricsApiHandler, error) {
	metricsApi := &MetricsApiHandler{
		options: options,
		network: n,
	}

	if value, found := options["pem"]; found {
		if p, ok := value.(string); ok {
			// This looks a little strange.  The yaml library used here does not handle multi-line strings properly.
			// To get things to parse correctly, the following must happen
			// 1: Remove BEING/END cert markers
			// 2: Remove all whitespace
			// 3: Restore the BEGIN/END cert markers
			//
			// Skipping step 1 will corrupt the markers when whitespace is removed
			regex := regexp.MustCompile(`-+\s*(BEGIN|END)\s+CERTIFICATE\s*-+`)
			p = string(regex.ReplaceAll([]byte(strings.TrimSpace(p)), []byte("")))

			t := "-----BEGIN CERTIFICATE-----\n" + strings.ReplaceAll(p, " ", "") + "\n-----END CERTIFICATE-----"

			block, _ := pem.Decode([]byte(t))
			if block == nil {
				err := errors.New("failed to parse metrics api PEM")
				return nil, err
			}
			cert, err := x509.ParseCertificate(block.Bytes)
			if err != nil {
				err := errors.New("failed to parse certificate: " + err.Error())
				return nil, err
			}
			metricsApi.scrapeCert = cert
		} else {
			return nil, errors.New("invalid configuration found for metrics pem.  The PEM must be a string")
		}
	} else {
		// TODO: Config location TBD
		pfxlog.Logger().Info("Metrics are enabled, however no PEM is provided.  Metric scrape identity will not be validated.  See TBD for instructions to set this up")
	}

	metricsApi.handler = metricsApi.newHandler()

	return metricsApi, nil
}

type MetricsApiHandler struct {
	handler     http.Handler
	network     *network.Network
	scrapeCert  *x509.Certificate
	options     map[interface{}]interface{}
	bindHandler channel.BindHandler
}

func (metricsApi *MetricsApiHandler) Binding() string {
	return MetricApiBinding
}

func (metricsApi *MetricsApiHandler) Options() map[interface{}]interface{} {
	return metricsApi.options
}

func (metricsApi *MetricsApiHandler) RootPath() string {
	// TODO:  MERGE BLOCKER: Requires fix to xweb demuxer to have a legit default handler.  The edge handler grabs everything except a hard-coded starts-with '/fabric'.
	return "/fabric-metrics"
}

func (metricsApi *MetricsApiHandler) IsHandler(r *http.Request) bool {
	return strings.HasPrefix(r.URL.Path, metricsApi.RootPath())
}

func (metricsApi *MetricsApiHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	metricsApi.handler.ServeHTTP(writer, request)
}

func (metricsApi *MetricsApiHandler) newHandler() http.Handler {

	handler := http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {

		if nil != metricsApi.scrapeCert {
			certOk := false
			for _, r := range r.TLS.PeerCertificates {
				if bytes.Equal(metricsApi.scrapeCert.Signature, r.Signature) {
					certOk = true
				}
			}

			if !certOk {
				rw.WriteHeader(http.StatusUnauthorized)
				return
			}
		}

		req, err := network.NewInspectionRequest(metricsApi.network, ".*", []string{"metrics:prometheus"})

		if err != nil {
			rw.Write([]byte(fmt.Sprintf("Failed to scrape metrics from %s:%s", metricsApi.network.GetAppId(), err.Error())))
			rw.WriteHeader(http.StatusInternalServerError)
			return
		}

		inspection := req.RunInspections()

		metricsResult, err := MapInspectResultToMetricsResult(inspection, "prometheus")

		if err != nil {
			rw.Write([]byte(fmt.Sprintf("Failed to convert metrics to prometheus format %s:%s", metricsApi.network.GetAppId(), err.Error())))
			rw.WriteHeader(http.StatusInternalServerError)
		} else {
			rw.Write([]byte(*metricsResult))
		}
	})

	return handler
}
