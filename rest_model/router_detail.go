// Code generated by go-swagger; DO NOT EDIT.

//
// Copyright NetFoundry, Inc.
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

package rest_model

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
	"github.com/go-openapi/validate"
)

// RouterDetail router detail
//
// swagger:model routerDetail
type RouterDetail struct {
	BaseEntity

	// connected
	// Required: true
	Connected *bool `json:"connected"`

	// cost
	// Required: true
	// Maximum: 65535
	// Minimum: 0
	Cost *int64 `json:"cost"`

	// fingerprint
	// Required: true
	Fingerprint *string `json:"fingerprint"`

	// listener address
	ListenerAddress string `json:"listenerAddress,omitempty"`

	// name
	// Required: true
	Name *string `json:"name"`

	// no traversal
	// Required: true
	NoTraversal *bool `json:"noTraversal"`

	// version info
	VersionInfo *VersionInfo `json:"versionInfo,omitempty"`
}

// UnmarshalJSON unmarshals this object from a JSON structure
func (m *RouterDetail) UnmarshalJSON(raw []byte) error {
	// AO0
	var aO0 BaseEntity
	if err := swag.ReadJSON(raw, &aO0); err != nil {
		return err
	}
	m.BaseEntity = aO0

	// AO1
	var dataAO1 struct {
		Connected *bool `json:"connected"`

		Cost *int64 `json:"cost"`

		Fingerprint *string `json:"fingerprint"`

		ListenerAddress string `json:"listenerAddress,omitempty"`

		Name *string `json:"name"`

		NoTraversal *bool `json:"noTraversal"`

		VersionInfo *VersionInfo `json:"versionInfo,omitempty"`
	}
	if err := swag.ReadJSON(raw, &dataAO1); err != nil {
		return err
	}

	m.Connected = dataAO1.Connected

	m.Cost = dataAO1.Cost

	m.Fingerprint = dataAO1.Fingerprint

	m.ListenerAddress = dataAO1.ListenerAddress

	m.Name = dataAO1.Name

	m.NoTraversal = dataAO1.NoTraversal

	m.VersionInfo = dataAO1.VersionInfo

	return nil
}

// MarshalJSON marshals this object to a JSON structure
func (m RouterDetail) MarshalJSON() ([]byte, error) {
	_parts := make([][]byte, 0, 2)

	aO0, err := swag.WriteJSON(m.BaseEntity)
	if err != nil {
		return nil, err
	}
	_parts = append(_parts, aO0)
	var dataAO1 struct {
		Connected *bool `json:"connected"`

		Cost *int64 `json:"cost"`

		Fingerprint *string `json:"fingerprint"`

		ListenerAddress string `json:"listenerAddress,omitempty"`

		Name *string `json:"name"`

		NoTraversal *bool `json:"noTraversal"`

		VersionInfo *VersionInfo `json:"versionInfo,omitempty"`
	}

	dataAO1.Connected = m.Connected

	dataAO1.Cost = m.Cost

	dataAO1.Fingerprint = m.Fingerprint

	dataAO1.ListenerAddress = m.ListenerAddress

	dataAO1.Name = m.Name

	dataAO1.NoTraversal = m.NoTraversal

	dataAO1.VersionInfo = m.VersionInfo

	jsonDataAO1, errAO1 := swag.WriteJSON(dataAO1)
	if errAO1 != nil {
		return nil, errAO1
	}
	_parts = append(_parts, jsonDataAO1)
	return swag.ConcatJSON(_parts...), nil
}

// Validate validates this router detail
func (m *RouterDetail) Validate(formats strfmt.Registry) error {
	var res []error

	// validation for a type composition with BaseEntity
	if err := m.BaseEntity.Validate(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateConnected(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateCost(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateFingerprint(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateName(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateNoTraversal(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateVersionInfo(formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *RouterDetail) validateConnected(formats strfmt.Registry) error {

	if err := validate.Required("connected", "body", m.Connected); err != nil {
		return err
	}

	return nil
}

func (m *RouterDetail) validateCost(formats strfmt.Registry) error {

	if err := validate.Required("cost", "body", m.Cost); err != nil {
		return err
	}

	if err := validate.MinimumInt("cost", "body", *m.Cost, 0, false); err != nil {
		return err
	}

	if err := validate.MaximumInt("cost", "body", *m.Cost, 65535, false); err != nil {
		return err
	}

	return nil
}

func (m *RouterDetail) validateFingerprint(formats strfmt.Registry) error {

	if err := validate.Required("fingerprint", "body", m.Fingerprint); err != nil {
		return err
	}

	return nil
}

func (m *RouterDetail) validateName(formats strfmt.Registry) error {

	if err := validate.Required("name", "body", m.Name); err != nil {
		return err
	}

	return nil
}

func (m *RouterDetail) validateNoTraversal(formats strfmt.Registry) error {

	if err := validate.Required("noTraversal", "body", m.NoTraversal); err != nil {
		return err
	}

	return nil
}

func (m *RouterDetail) validateVersionInfo(formats strfmt.Registry) error {

	if swag.IsZero(m.VersionInfo) { // not required
		return nil
	}

	if m.VersionInfo != nil {
		if err := m.VersionInfo.Validate(formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("versionInfo")
			} else if ce, ok := err.(*errors.CompositeError); ok {
				return ce.ValidateName("versionInfo")
			}
			return err
		}
	}

	return nil
}

// ContextValidate validate this router detail based on the context it is used
func (m *RouterDetail) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
	var res []error

	// validation for a type composition with BaseEntity
	if err := m.BaseEntity.ContextValidate(ctx, formats); err != nil {
		res = append(res, err)
	}

	if err := m.contextValidateVersionInfo(ctx, formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *RouterDetail) contextValidateVersionInfo(ctx context.Context, formats strfmt.Registry) error {

	if m.VersionInfo != nil {
		if err := m.VersionInfo.ContextValidate(ctx, formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("versionInfo")
			} else if ce, ok := err.(*errors.CompositeError); ok {
				return ce.ValidateName("versionInfo")
			}
			return err
		}
	}

	return nil
}

// MarshalBinary interface implementation
func (m *RouterDetail) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *RouterDetail) UnmarshalBinary(b []byte) error {
	var res RouterDetail
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
