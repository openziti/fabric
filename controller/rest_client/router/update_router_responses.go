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

package router

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"io"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"

	"github.com/openziti/fabric/controller/rest_model"
)

// UpdateRouterReader is a Reader for the UpdateRouter structure.
type UpdateRouterReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *UpdateRouterReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewUpdateRouterOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	case 400:
		result := NewUpdateRouterBadRequest()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 401:
		result := NewUpdateRouterUnauthorized()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 404:
		result := NewUpdateRouterNotFound()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	default:
		return nil, runtime.NewAPIError("response status code does not match any response statuses defined for this endpoint in the swagger spec", response, response.Code())
	}
}

// NewUpdateRouterOK creates a UpdateRouterOK with default headers values
func NewUpdateRouterOK() *UpdateRouterOK {
	return &UpdateRouterOK{}
}

/*
UpdateRouterOK describes a response with status code 200, with default header values.

The update request was successful and the resource has been altered
*/
type UpdateRouterOK struct {
	Payload *rest_model.Empty
}

// IsSuccess returns true when this update router o k response has a 2xx status code
func (o *UpdateRouterOK) IsSuccess() bool {
	return true
}

// IsRedirect returns true when this update router o k response has a 3xx status code
func (o *UpdateRouterOK) IsRedirect() bool {
	return false
}

// IsClientError returns true when this update router o k response has a 4xx status code
func (o *UpdateRouterOK) IsClientError() bool {
	return false
}

// IsServerError returns true when this update router o k response has a 5xx status code
func (o *UpdateRouterOK) IsServerError() bool {
	return false
}

// IsCode returns true when this update router o k response a status code equal to that given
func (o *UpdateRouterOK) IsCode(code int) bool {
	return code == 200
}

// Code gets the status code for the update router o k response
func (o *UpdateRouterOK) Code() int {
	return 200
}

func (o *UpdateRouterOK) Error() string {
	return fmt.Sprintf("[PUT /routers/{id}][%d] updateRouterOK  %+v", 200, o.Payload)
}

func (o *UpdateRouterOK) String() string {
	return fmt.Sprintf("[PUT /routers/{id}][%d] updateRouterOK  %+v", 200, o.Payload)
}

func (o *UpdateRouterOK) GetPayload() *rest_model.Empty {
	return o.Payload
}

func (o *UpdateRouterOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(rest_model.Empty)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewUpdateRouterBadRequest creates a UpdateRouterBadRequest with default headers values
func NewUpdateRouterBadRequest() *UpdateRouterBadRequest {
	return &UpdateRouterBadRequest{}
}

/*
UpdateRouterBadRequest describes a response with status code 400, with default header values.

The supplied request contains invalid fields or could not be parsed (json and non-json bodies). The error's code, message, and cause fields can be inspected for further information
*/
type UpdateRouterBadRequest struct {
	Payload *rest_model.APIErrorEnvelope
}

// IsSuccess returns true when this update router bad request response has a 2xx status code
func (o *UpdateRouterBadRequest) IsSuccess() bool {
	return false
}

// IsRedirect returns true when this update router bad request response has a 3xx status code
func (o *UpdateRouterBadRequest) IsRedirect() bool {
	return false
}

// IsClientError returns true when this update router bad request response has a 4xx status code
func (o *UpdateRouterBadRequest) IsClientError() bool {
	return true
}

// IsServerError returns true when this update router bad request response has a 5xx status code
func (o *UpdateRouterBadRequest) IsServerError() bool {
	return false
}

// IsCode returns true when this update router bad request response a status code equal to that given
func (o *UpdateRouterBadRequest) IsCode(code int) bool {
	return code == 400
}

// Code gets the status code for the update router bad request response
func (o *UpdateRouterBadRequest) Code() int {
	return 400
}

func (o *UpdateRouterBadRequest) Error() string {
	return fmt.Sprintf("[PUT /routers/{id}][%d] updateRouterBadRequest  %+v", 400, o.Payload)
}

func (o *UpdateRouterBadRequest) String() string {
	return fmt.Sprintf("[PUT /routers/{id}][%d] updateRouterBadRequest  %+v", 400, o.Payload)
}

func (o *UpdateRouterBadRequest) GetPayload() *rest_model.APIErrorEnvelope {
	return o.Payload
}

func (o *UpdateRouterBadRequest) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(rest_model.APIErrorEnvelope)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewUpdateRouterUnauthorized creates a UpdateRouterUnauthorized with default headers values
func NewUpdateRouterUnauthorized() *UpdateRouterUnauthorized {
	return &UpdateRouterUnauthorized{}
}

/*
UpdateRouterUnauthorized describes a response with status code 401, with default header values.

The currently supplied session does not have the correct access rights to request this resource
*/
type UpdateRouterUnauthorized struct {
	Payload *rest_model.APIErrorEnvelope
}

// IsSuccess returns true when this update router unauthorized response has a 2xx status code
func (o *UpdateRouterUnauthorized) IsSuccess() bool {
	return false
}

// IsRedirect returns true when this update router unauthorized response has a 3xx status code
func (o *UpdateRouterUnauthorized) IsRedirect() bool {
	return false
}

// IsClientError returns true when this update router unauthorized response has a 4xx status code
func (o *UpdateRouterUnauthorized) IsClientError() bool {
	return true
}

// IsServerError returns true when this update router unauthorized response has a 5xx status code
func (o *UpdateRouterUnauthorized) IsServerError() bool {
	return false
}

// IsCode returns true when this update router unauthorized response a status code equal to that given
func (o *UpdateRouterUnauthorized) IsCode(code int) bool {
	return code == 401
}

// Code gets the status code for the update router unauthorized response
func (o *UpdateRouterUnauthorized) Code() int {
	return 401
}

func (o *UpdateRouterUnauthorized) Error() string {
	return fmt.Sprintf("[PUT /routers/{id}][%d] updateRouterUnauthorized  %+v", 401, o.Payload)
}

func (o *UpdateRouterUnauthorized) String() string {
	return fmt.Sprintf("[PUT /routers/{id}][%d] updateRouterUnauthorized  %+v", 401, o.Payload)
}

func (o *UpdateRouterUnauthorized) GetPayload() *rest_model.APIErrorEnvelope {
	return o.Payload
}

func (o *UpdateRouterUnauthorized) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(rest_model.APIErrorEnvelope)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewUpdateRouterNotFound creates a UpdateRouterNotFound with default headers values
func NewUpdateRouterNotFound() *UpdateRouterNotFound {
	return &UpdateRouterNotFound{}
}

/*
UpdateRouterNotFound describes a response with status code 404, with default header values.

The requested resource does not exist
*/
type UpdateRouterNotFound struct {
	Payload *rest_model.APIErrorEnvelope
}

// IsSuccess returns true when this update router not found response has a 2xx status code
func (o *UpdateRouterNotFound) IsSuccess() bool {
	return false
}

// IsRedirect returns true when this update router not found response has a 3xx status code
func (o *UpdateRouterNotFound) IsRedirect() bool {
	return false
}

// IsClientError returns true when this update router not found response has a 4xx status code
func (o *UpdateRouterNotFound) IsClientError() bool {
	return true
}

// IsServerError returns true when this update router not found response has a 5xx status code
func (o *UpdateRouterNotFound) IsServerError() bool {
	return false
}

// IsCode returns true when this update router not found response a status code equal to that given
func (o *UpdateRouterNotFound) IsCode(code int) bool {
	return code == 404
}

// Code gets the status code for the update router not found response
func (o *UpdateRouterNotFound) Code() int {
	return 404
}

func (o *UpdateRouterNotFound) Error() string {
	return fmt.Sprintf("[PUT /routers/{id}][%d] updateRouterNotFound  %+v", 404, o.Payload)
}

func (o *UpdateRouterNotFound) String() string {
	return fmt.Sprintf("[PUT /routers/{id}][%d] updateRouterNotFound  %+v", 404, o.Payload)
}

func (o *UpdateRouterNotFound) GetPayload() *rest_model.APIErrorEnvelope {
	return o.Payload
}

func (o *UpdateRouterNotFound) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(rest_model.APIErrorEnvelope)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}