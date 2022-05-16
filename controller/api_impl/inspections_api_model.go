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
	"encoding/json"
	"github.com/openziti/fabric/controller/network"
	"github.com/openziti/fabric/rest_model"
	"strings"
)

const EntityNameInspect = "inspections"

// Maps individual response from inspection into overall inspection result
func MapInspectResultToRestModel(inspectResult *network.InspectResult) *rest_model.InspectResponse {
	resp := &rest_model.InspectResponse{
		Errors:  inspectResult.Errors,
		Success: &inspectResult.Success,
	}
	for _, val := range inspectResult.Results {
		var emitVal interface{}
		// TODO:  Check for metrics.  If metrics,  then convert to metrics PB,  then use metrics adapter to marshal to stream of metrics events.   Marshal stream of metrics events into an array of metrics to be returned.
		// Metrics Msg -> []metrics event -> Marshalled json array
		//if val.Name == "metrics" {
		//	msg := &metrics_pb.MetricsMessage{}
		//	if err := json.Unmarshal([]byte(val.Value), msg); err == nil {
		//
		//		emitVal = mapVal
		//	}
		//	val.Value
		//} else {
		if strings.HasPrefix(val.Value, "{") {
			mapVal := map[string]interface{}{}
			if err := json.Unmarshal([]byte(val.Value), &mapVal); err == nil {
				emitVal = mapVal
			}
		} else if strings.HasPrefix(val.Value, "[") {
			var arrayVal []interface{}
			if err := json.Unmarshal([]byte(val.Value), &arrayVal); err == nil {
				emitVal = arrayVal
			}
		}
		//}

		if emitVal == nil {
			emitVal = val.Value
		}

		resp.Values = append(resp.Values, &rest_model.InspectResponseValue{
			AppID: &val.AppId,
			Name:  &val.Name,
			Value: emitVal,
		})
	}
	return resp
}
