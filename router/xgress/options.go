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

package xgress

import (
	"encoding/json"
	"github.com/pkg/errors"
	"time"
)

// Options contains common Xgress configuration options
type Options struct {
	Mtu               int32
	Retransmission    bool
	RandomDrops       bool
	Drop1InN          int32
	GetSessionTimeout time.Duration
}

func LoadOptions(data OptionsData) (*Options, error) {
	options := DefaultOptions()

	if value, found := data["options"]; found {
		data = value.(map[interface{}]interface{})

		if value, found := data["mtu"]; found {
			options.Mtu = int32(value.(int))
		}
		if value, found := data["retransmission"]; found {
			options.Retransmission = value.(bool)
		}
		if value, found := data["randomDrops"]; found {
			options.RandomDrops = value.(bool)
		}
		if value, found := data["drop1InN"]; found {
			options.Drop1InN = int32(value.(int))
		}

		if value, found := data["getSessionTimeout"]; found {
			getSessionTimeout, err := time.ParseDuration(value.(string))
			if err != nil {
				return nil, errors.Wrap(err, "invalid 'getSessionTimeout' value")
			}
			options.GetSessionTimeout = getSessionTimeout
		}
	}

	return options, nil
}

func DefaultOptions() *Options {
	return &Options{
		Mtu:               64 * 1024,
		Retransmission:    true,
		RandomDrops:       false,
		Drop1InN:          100,
		GetSessionTimeout: 30 * time.Second,
	}
}

func (options Options) String() string {
	data, err := json.Marshal(options)
	if err != nil {
		return err.Error()
	}
	return string(data)
}
