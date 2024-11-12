// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
	"github.com/go-openapi/validate"
)

// MetricsConfig metrics config
//
// swagger:model MetricsConfig
type MetricsConfig struct {

	// value for prometheus enable or not
	// Required: true
	Prometheus *bool `json:"prometheus"`
}

// Validate validates this metrics config
func (m *MetricsConfig) Validate(formats strfmt.Registry) error {
	var res []error

	if err := m.validatePrometheus(formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *MetricsConfig) validatePrometheus(formats strfmt.Registry) error {

	if err := validate.Required("prometheus", "body", m.Prometheus); err != nil {
		return err
	}

	return nil
}

// ContextValidate validates this metrics config based on context it is used
func (m *MetricsConfig) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (m *MetricsConfig) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *MetricsConfig) UnmarshalBinary(b []byte) error {
	var res MetricsConfig
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}