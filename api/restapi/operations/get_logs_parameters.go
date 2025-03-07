// Code generated by go-swagger; DO NOT EDIT.

package operations

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"
)

// NewGetLogsParams creates a new GetLogsParams object
//
// There are no default values defined in the spec.
func NewGetLogsParams() GetLogsParams {

	return GetLogsParams{}
}

// GetLogsParams contains all the bound params for the get logs operation
// typically these are obtained from a http.Request
//
// swagger:parameters GetLogs
type GetLogsParams struct {

	// HTTP Request Object
	HTTPRequest *http.Request `json:"-"`

	/*Filter logs containing a specific keyword or phrase.
	  In: query
	*/
	Keyword *string
	/*Filter logs by level (e.g., INFO, ERROR, DEBUG).
	  In: query
	*/
	Level *string
	/*Number of log lines to fetch (default is 100).
	  In: query
	*/
	Lines *string
}

// BindRequest both binds and validates a request, it assumes that complex things implement a Validatable(strfmt.Registry) error interface
// for simple values it will use straight method calls.
//
// To ensure default values, the struct must have been initialized with NewGetLogsParams() beforehand.
func (o *GetLogsParams) BindRequest(r *http.Request, route *middleware.MatchedRoute) error {
	var res []error

	o.HTTPRequest = r

	qs := runtime.Values(r.URL.Query())

	qKeyword, qhkKeyword, _ := qs.GetOK("keyword")
	if err := o.bindKeyword(qKeyword, qhkKeyword, route.Formats); err != nil {
		res = append(res, err)
	}

	qLevel, qhkLevel, _ := qs.GetOK("level")
	if err := o.bindLevel(qLevel, qhkLevel, route.Formats); err != nil {
		res = append(res, err)
	}

	qLines, qhkLines, _ := qs.GetOK("lines")
	if err := o.bindLines(qLines, qhkLines, route.Formats); err != nil {
		res = append(res, err)
	}
	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

// bindKeyword binds and validates parameter Keyword from query.
func (o *GetLogsParams) bindKeyword(rawData []string, hasKey bool, formats strfmt.Registry) error {
	var raw string
	if len(rawData) > 0 {
		raw = rawData[len(rawData)-1]
	}

	// Required: false
	// AllowEmptyValue: false

	if raw == "" { // empty values pass all other validations
		return nil
	}
	o.Keyword = &raw

	return nil
}

// bindLevel binds and validates parameter Level from query.
func (o *GetLogsParams) bindLevel(rawData []string, hasKey bool, formats strfmt.Registry) error {
	var raw string
	if len(rawData) > 0 {
		raw = rawData[len(rawData)-1]
	}

	// Required: false
	// AllowEmptyValue: false

	if raw == "" { // empty values pass all other validations
		return nil
	}
	o.Level = &raw

	return nil
}

// bindLines binds and validates parameter Lines from query.
func (o *GetLogsParams) bindLines(rawData []string, hasKey bool, formats strfmt.Registry) error {
	var raw string
	if len(rawData) > 0 {
		raw = rawData[len(rawData)-1]
	}

	// Required: false
	// AllowEmptyValue: false

	if raw == "" { // empty values pass all other validations
		return nil
	}
	o.Lines = &raw

	return nil
}
