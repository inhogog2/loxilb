// Code generated by go-swagger; DO NOT EDIT.

package operations

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the generate command

import (
	"errors"
	"net/url"
	golangswaggerpaths "path"
	"strings"

	"github.com/go-openapi/swag"
)

// DeleteConfigLoadbalancerExternalipaddressIPAddressPortPortPortmaxPortmaxProtocolProtoURL generates an URL for the delete config loadbalancer externalipaddress IP address port port portmax portmax protocol proto operation
type DeleteConfigLoadbalancerExternalipaddressIPAddressPortPortPortmaxPortmaxProtocolProtoURL struct {
	IPAddress string
	Port      float64
	Portmax   float64
	Proto     string

	Bgp   *bool
	Block *float64

	_basePath string
	// avoid unkeyed usage
	_ struct{}
}

// WithBasePath sets the base path for this url builder, only required when it's different from the
// base path specified in the swagger spec.
// When the value of the base path is an empty string
func (o *DeleteConfigLoadbalancerExternalipaddressIPAddressPortPortPortmaxPortmaxProtocolProtoURL) WithBasePath(bp string) *DeleteConfigLoadbalancerExternalipaddressIPAddressPortPortPortmaxPortmaxProtocolProtoURL {
	o.SetBasePath(bp)
	return o
}

// SetBasePath sets the base path for this url builder, only required when it's different from the
// base path specified in the swagger spec.
// When the value of the base path is an empty string
func (o *DeleteConfigLoadbalancerExternalipaddressIPAddressPortPortPortmaxPortmaxProtocolProtoURL) SetBasePath(bp string) {
	o._basePath = bp
}

// Build a url path and query string
func (o *DeleteConfigLoadbalancerExternalipaddressIPAddressPortPortPortmaxPortmaxProtocolProtoURL) Build() (*url.URL, error) {
	var _result url.URL

	var _path = "/config/loadbalancer/externalipaddress/{ip_address}/port/{port}/portmax/{portmax}/protocol/{proto}"

	iPAddress := o.IPAddress
	if iPAddress != "" {
		_path = strings.Replace(_path, "{ip_address}", iPAddress, -1)
	} else {
		return nil, errors.New("ipAddress is required on DeleteConfigLoadbalancerExternalipaddressIPAddressPortPortPortmaxPortmaxProtocolProtoURL")
	}

	port := swag.FormatFloat64(o.Port)
	if port != "" {
		_path = strings.Replace(_path, "{port}", port, -1)
	} else {
		return nil, errors.New("port is required on DeleteConfigLoadbalancerExternalipaddressIPAddressPortPortPortmaxPortmaxProtocolProtoURL")
	}

	portmax := swag.FormatFloat64(o.Portmax)
	if portmax != "" {
		_path = strings.Replace(_path, "{portmax}", portmax, -1)
	} else {
		return nil, errors.New("portmax is required on DeleteConfigLoadbalancerExternalipaddressIPAddressPortPortPortmaxPortmaxProtocolProtoURL")
	}

	proto := o.Proto
	if proto != "" {
		_path = strings.Replace(_path, "{proto}", proto, -1)
	} else {
		return nil, errors.New("proto is required on DeleteConfigLoadbalancerExternalipaddressIPAddressPortPortPortmaxPortmaxProtocolProtoURL")
	}

	_basePath := o._basePath
	if _basePath == "" {
		_basePath = "/netlox/v1"
	}
	_result.Path = golangswaggerpaths.Join(_basePath, _path)

	qs := make(url.Values)

	var bgpQ string
	if o.Bgp != nil {
		bgpQ = swag.FormatBool(*o.Bgp)
	}
	if bgpQ != "" {
		qs.Set("bgp", bgpQ)
	}

	var blockQ string
	if o.Block != nil {
		blockQ = swag.FormatFloat64(*o.Block)
	}
	if blockQ != "" {
		qs.Set("block", blockQ)
	}

	_result.RawQuery = qs.Encode()

	return &_result, nil
}

// Must is a helper function to panic when the url builder returns an error
func (o *DeleteConfigLoadbalancerExternalipaddressIPAddressPortPortPortmaxPortmaxProtocolProtoURL) Must(u *url.URL, err error) *url.URL {
	if err != nil {
		panic(err)
	}
	if u == nil {
		panic("url can't be nil")
	}
	return u
}

// String returns the string representation of the path with query string
func (o *DeleteConfigLoadbalancerExternalipaddressIPAddressPortPortPortmaxPortmaxProtocolProtoURL) String() string {
	return o.Must(o.Build()).String()
}

// BuildFull builds a full url with scheme, host, path and query string
func (o *DeleteConfigLoadbalancerExternalipaddressIPAddressPortPortPortmaxPortmaxProtocolProtoURL) BuildFull(scheme, host string) (*url.URL, error) {
	if scheme == "" {
		return nil, errors.New("scheme is required for a full url on DeleteConfigLoadbalancerExternalipaddressIPAddressPortPortPortmaxPortmaxProtocolProtoURL")
	}
	if host == "" {
		return nil, errors.New("host is required for a full url on DeleteConfigLoadbalancerExternalipaddressIPAddressPortPortPortmaxPortmaxProtocolProtoURL")
	}

	base, err := o.Build()
	if err != nil {
		return nil, err
	}

	base.Scheme = scheme
	base.Host = host
	return base, nil
}

// StringFull returns the string representation of a complete url
func (o *DeleteConfigLoadbalancerExternalipaddressIPAddressPortPortPortmaxPortmaxProtocolProtoURL) StringFull(scheme, host string) string {
	return o.Must(o.BuildFull(scheme, host)).String()
}
