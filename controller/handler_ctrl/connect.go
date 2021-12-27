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

package handler_ctrl

import (
	"crypto/sha1"
	"crypto/x509"
	"fmt"
	"github.com/michaelquigley/pfxlog"
	"github.com/openziti/fabric/controller/network"
	"github.com/openziti/fabric/controller/xctrl"
	"github.com/openziti/foundation/channel2"
	"github.com/openziti/foundation/identity/identity"
	"github.com/openziti/foundation/util/errorz"
	"github.com/pkg/errors"
)

type ConnectHandler struct {
	identity identity.Identity
	network  *network.Network
	xctrls   []xctrl.Xctrl
}

func NewConnectHandler(identity identity.Identity, network *network.Network, xctrls []xctrl.Xctrl) *ConnectHandler {
	return &ConnectHandler{
		identity: identity,
		network:  network,
		xctrls:   xctrls,
	}
}

func (self *ConnectHandler) HandleConnection(hello *channel2.Hello, certificates []*x509.Certificate) error {
	id := hello.IdToken

	// verify cert chain
	if len(certificates) == 0 {
		return errors.Errorf("no certificates provided, unable to verify dialer, routerId: %v", id)
	}

	config := self.identity.ServerTLSConfig()

	opts := x509.VerifyOptions{
		Roots:         config.RootCAs,
		Intermediates: x509.NewCertPool(),
		KeyUsages:     []x509.ExtKeyUsage{x509.ExtKeyUsageAny},
	}

	var errorList errorz.MultipleErrors

	for _, cert := range certificates {
		if _, err := cert.Verify(opts); err == nil {
			return nil
		} else {
			errorList = append(errorList, err)
		}
	}

	if len(errorList) > 0 {
		return errorList.ToError()
	}

	log := pfxlog.Logger().WithField("routerId", hello.IdToken)

	fingerprint := ""
	if len(certificates) > 0 {
		log.Debugf("peer has [%d] certificates", len(certificates))
		for i, c := range certificates {
			fingerprint = fmt.Sprintf("%x", sha1.Sum(c.Raw))
			log.Debugf("%d): peer certificate fingerprint [%s]", i, fingerprint)
			log.Debugf("%d): peer common name [%s]", i, c.Subject.CommonName)
		}
	} else {
		log.Warn("peer has no certificates")
	}

	if self.network.ConnectedRouter(id) {
		router := self.network.GetConnectedRouter(id)
		name := "unknown"
		if router != nil {
			name = router.Name
		}
		log.WithField("routerName", name).Error("router already connected")
		return fmt.Errorf("router already connected id: %s, name: %s", id, name)
	}

	if r, err := self.network.GetRouter(id); err == nil {
		if r.Fingerprint == nil {
			log.Error("router enrollment incomplete")
			return errors.Errorf("router enrollment incomplete, routerId: %v", id)
		}
		if *r.Fingerprint != fingerprint {
			log.WithField("fp", *r.Fingerprint).WithField("givenFp", fingerprint).Error("router fingerprint mismatch")
			return errors.Errorf("incorrect fingerprint/unenrolled router, routerId: %v, given fingerprint: %v", id, fingerprint)
		}
	} else {
		log.Error("unknown/unenrolled router")
		return errors.Errorf("unknown/unenrolled router, routerId: %v", id)
	}

	return nil
}
