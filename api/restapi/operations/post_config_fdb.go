// Code generated by go-swagger; DO NOT EDIT.

package operations

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the generate command

import (
	"net/http"

	"github.com/go-openapi/runtime/middleware"
)

// PostConfigFdbHandlerFunc turns a function with the right signature into a post config fdb handler
type PostConfigFdbHandlerFunc func(PostConfigFdbParams, interface{}) middleware.Responder

// Handle executing the request and returning a response
func (fn PostConfigFdbHandlerFunc) Handle(params PostConfigFdbParams, principal interface{}) middleware.Responder {
	return fn(params, principal)
}

// PostConfigFdbHandler interface for that can handle valid post config fdb params
type PostConfigFdbHandler interface {
	Handle(PostConfigFdbParams, interface{}) middleware.Responder
}

// NewPostConfigFdb creates a new http.Handler for the post config fdb operation
func NewPostConfigFdb(ctx *middleware.Context, handler PostConfigFdbHandler) *PostConfigFdb {
	return &PostConfigFdb{Context: ctx, Handler: handler}
}

/*
	PostConfigFdb swagger:route POST /config/fdb postConfigFdb

# Assign FDB in the device

Assign FDB in the device
*/
type PostConfigFdb struct {
	Context *middleware.Context
	Handler PostConfigFdbHandler
}

func (o *PostConfigFdb) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	route, rCtx, _ := o.Context.RouteInfo(r)
	if rCtx != nil {
		*r = *rCtx
	}
	var Params = NewPostConfigFdbParams()
	uprinc, aCtx, err := o.Context.Authorize(r, route)
	if err != nil {
		o.Context.Respond(rw, r, route.Produces, route, err)
		return
	}
	if aCtx != nil {
		*r = *aCtx
	}
	var principal interface{}
	if uprinc != nil {
		principal = uprinc.(interface{}) // this is really a interface{}, I promise
	}

	if err := o.Context.BindValidRequest(r, route, &Params); err != nil { // bind params
		o.Context.Respond(rw, r, route.Produces, route, err)
		return
	}

	res := o.Handler.Handle(Params, principal) // actually handle the request
	o.Context.Respond(rw, r, route.Produces, route, res)

}
