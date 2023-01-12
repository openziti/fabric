// Code generated by go-swagger; DO NOT EDIT.

//
// Copyright NetFoundry Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// __          __              _
// \ \        / /             (_)
//  \ \  /\  / /_ _ _ __ _ __  _ _ __   __ _
//   \ \/  \/ / _` | '__| '_ \| | '_ \ / _` |
//    \  /\  / (_| | |  | | | | | | | | (_| | : This file is generated, do not edit it.
//     \/  \/ \__,_|_|  |_| |_|_|_| |_|\__, |
//                                      __/ |
//                                     |___/

package rest_client

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"github.com/go-openapi/runtime"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"

	"github.com/openziti/fabric/rest_client/circuit"
	"github.com/openziti/fabric/rest_client/database"
	"github.com/openziti/fabric/rest_client/inspect"
	"github.com/openziti/fabric/rest_client/link"
	"github.com/openziti/fabric/rest_client/raft"
	"github.com/openziti/fabric/rest_client/router"
	"github.com/openziti/fabric/rest_client/service"
	"github.com/openziti/fabric/rest_client/terminator"
)

// Default ziti fabric HTTP client.
var Default = NewHTTPClient(nil)

const (
	// DefaultHost is the default Host
	// found in Meta (info) section of spec file
	DefaultHost string = "demo.ziti.dev"
	// DefaultBasePath is the default BasePath
	// found in Meta (info) section of spec file
	DefaultBasePath string = "/fabric/v1"
)

// DefaultSchemes are the default schemes found in Meta (info) section of spec file
var DefaultSchemes = []string{"https"}

// NewHTTPClient creates a new ziti fabric HTTP client.
func NewHTTPClient(formats strfmt.Registry) *ZitiFabric {
	return NewHTTPClientWithConfig(formats, nil)
}

// NewHTTPClientWithConfig creates a new ziti fabric HTTP client,
// using a customizable transport config.
func NewHTTPClientWithConfig(formats strfmt.Registry, cfg *TransportConfig) *ZitiFabric {
	// ensure nullable parameters have default
	if cfg == nil {
		cfg = DefaultTransportConfig()
	}

	// create transport and client
	transport := httptransport.New(cfg.Host, cfg.BasePath, cfg.Schemes)
	return New(transport, formats)
}

// New creates a new ziti fabric client
func New(transport runtime.ClientTransport, formats strfmt.Registry) *ZitiFabric {
	// ensure nullable parameters have default
	if formats == nil {
		formats = strfmt.Default
	}

	cli := new(ZitiFabric)
	cli.Transport = transport
	cli.Circuit = circuit.New(transport, formats)
	cli.Database = database.New(transport, formats)
	cli.Inspect = inspect.New(transport, formats)
	cli.Link = link.New(transport, formats)
	cli.Raft = raft.New(transport, formats)
	cli.Router = router.New(transport, formats)
	cli.Service = service.New(transport, formats)
	cli.Terminator = terminator.New(transport, formats)
	return cli
}

// DefaultTransportConfig creates a TransportConfig with the
// default settings taken from the meta section of the spec file.
func DefaultTransportConfig() *TransportConfig {
	return &TransportConfig{
		Host:     DefaultHost,
		BasePath: DefaultBasePath,
		Schemes:  DefaultSchemes,
	}
}

// TransportConfig contains the transport related info,
// found in the meta section of the spec file.
type TransportConfig struct {
	Host     string
	BasePath string
	Schemes  []string
}

// WithHost overrides the default host,
// provided by the meta section of the spec file.
func (cfg *TransportConfig) WithHost(host string) *TransportConfig {
	cfg.Host = host
	return cfg
}

// WithBasePath overrides the default basePath,
// provided by the meta section of the spec file.
func (cfg *TransportConfig) WithBasePath(basePath string) *TransportConfig {
	cfg.BasePath = basePath
	return cfg
}

// WithSchemes overrides the default schemes,
// provided by the meta section of the spec file.
func (cfg *TransportConfig) WithSchemes(schemes []string) *TransportConfig {
	cfg.Schemes = schemes
	return cfg
}

// ZitiFabric is a client for ziti fabric
type ZitiFabric struct {
	Circuit circuit.ClientService

	Database database.ClientService

	Inspect inspect.ClientService

	Link link.ClientService

	Raft raft.ClientService

	Router router.ClientService

	Service service.ClientService

	Terminator terminator.ClientService

	Transport runtime.ClientTransport
}

// SetTransport changes the transport on the client and all its subresources
func (c *ZitiFabric) SetTransport(transport runtime.ClientTransport) {
	c.Transport = transport
	c.Circuit.SetTransport(transport)
	c.Database.SetTransport(transport)
	c.Inspect.SetTransport(transport)
	c.Link.SetTransport(transport)
	c.Raft.SetTransport(transport)
	c.Router.SetTransport(transport)
	c.Service.SetTransport(transport)
	c.Terminator.SetTransport(transport)
}
