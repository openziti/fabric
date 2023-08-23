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

package circuit

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"io"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"

	"github.com/openziti/fabric/controller/rest_model"
)

// ListCircuitsReader is a Reader for the ListCircuits structure.
type ListCircuitsReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *ListCircuitsReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewListCircuitsOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	case 401:
		result := NewListCircuitsUnauthorized()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	default:
		return nil, runtime.NewAPIError("response status code does not match any response statuses defined for this endpoint in the swagger spec", response, response.Code())
	}
}

// NewListCircuitsOK creates a ListCircuitsOK with default headers values
func NewListCircuitsOK() *ListCircuitsOK {
	return &ListCircuitsOK{}
}

/*
ListCircuitsOK describes a response with status code 200, with default header values.

A list of circuits
*/
type ListCircuitsOK struct {
	Payload *rest_model.ListCircuitsEnvelope
}

// IsSuccess returns true when this list circuits o k response has a 2xx status code
func (o *ListCircuitsOK) IsSuccess() bool {
	return true
}

// IsRedirect returns true when this list circuits o k response has a 3xx status code
func (o *ListCircuitsOK) IsRedirect() bool {
	return false
}

// IsClientError returns true when this list circuits o k response has a 4xx status code
func (o *ListCircuitsOK) IsClientError() bool {
	return false
}

// IsServerError returns true when this list circuits o k response has a 5xx status code
func (o *ListCircuitsOK) IsServerError() bool {
	return false
}

// IsCode returns true when this list circuits o k response a status code equal to that given
func (o *ListCircuitsOK) IsCode(code int) bool {
	return code == 200
}

// Code gets the status code for the list circuits o k response
func (o *ListCircuitsOK) Code() int {
	return 200
}

func (o *ListCircuitsOK) Error() string {
	return fmt.Sprintf("[GET /circuits][%d] listCircuitsOK  %+v", 200, o.Payload)
}

func (o *ListCircuitsOK) String() string {
	return fmt.Sprintf("[GET /circuits][%d] listCircuitsOK  %+v", 200, o.Payload)
}

func (o *ListCircuitsOK) GetPayload() *rest_model.ListCircuitsEnvelope {
	return o.Payload
}

func (o *ListCircuitsOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(rest_model.ListCircuitsEnvelope)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewListCircuitsUnauthorized creates a ListCircuitsUnauthorized with default headers values
func NewListCircuitsUnauthorized() *ListCircuitsUnauthorized {
	return &ListCircuitsUnauthorized{}
}

/*
ListCircuitsUnauthorized describes a response with status code 401, with default header values.

The currently supplied session does not have the correct access rights to request this resource
*/
type ListCircuitsUnauthorized struct {
	Payload *rest_model.APIErrorEnvelope
}

// IsSuccess returns true when this list circuits unauthorized response has a 2xx status code
func (o *ListCircuitsUnauthorized) IsSuccess() bool {
	return false
}

// IsRedirect returns true when this list circuits unauthorized response has a 3xx status code
func (o *ListCircuitsUnauthorized) IsRedirect() bool {
	return false
}

// IsClientError returns true when this list circuits unauthorized response has a 4xx status code
func (o *ListCircuitsUnauthorized) IsClientError() bool {
	return true
}

// IsServerError returns true when this list circuits unauthorized response has a 5xx status code
func (o *ListCircuitsUnauthorized) IsServerError() bool {
	return false
}

// IsCode returns true when this list circuits unauthorized response a status code equal to that given
func (o *ListCircuitsUnauthorized) IsCode(code int) bool {
	return code == 401
}

// Code gets the status code for the list circuits unauthorized response
func (o *ListCircuitsUnauthorized) Code() int {
	return 401
}

func (o *ListCircuitsUnauthorized) Error() string {
	return fmt.Sprintf("[GET /circuits][%d] listCircuitsUnauthorized  %+v", 401, o.Payload)
}

func (o *ListCircuitsUnauthorized) String() string {
	return fmt.Sprintf("[GET /circuits][%d] listCircuitsUnauthorized  %+v", 401, o.Payload)
}

func (o *ListCircuitsUnauthorized) GetPayload() *rest_model.APIErrorEnvelope {
	return o.Payload
}

func (o *ListCircuitsUnauthorized) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(rest_model.APIErrorEnvelope)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}