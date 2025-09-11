/*
 * Copyright (c) 2022 NetLOX Inc
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at:
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package handler

import (
	"github.com/loxilb-io/loxilb/api/models"

	"github.com/go-openapi/runtime/middleware"
	"github.com/loxilb-io/loxilb/api/loxinlp"
	"github.com/loxilb-io/loxilb/api/restapi/operations"
	tk "github.com/loxilb-io/loxilib"
)

func ConfigPostNeighbor(params operations.PostConfigNeighborParams, principal interface{}) middleware.Responder {
	tk.LogIt(tk.LogTrace, "api: IPv4 Neighbor %s API called. url : %s\n", params.HTTPRequest.Method, params.HTTPRequest.URL)
	if params.Attr.IPAddress == nil || params.Attr.Dev == nil || params.Attr.MacAddress == nil {
		return &ErrorResponse{Payload: ResultErrorResponseErrorMessage("invalid parameters")}
	}
	err := loxinlp.AddNeighNoHook(*params.Attr.IPAddress, *params.Attr.Dev, *params.Attr.MacAddress)
	if err != nil {
		return &ErrorResponse{Payload: ResultErrorResponseErrorMessage(err.Error())}
	}
	return &ResultResponse{Result: "Success"}
}

func ConfigDeleteNeighbor(params operations.DeleteConfigNeighborIPAddressDevIfNameParams, principal interface{}) middleware.Responder {
	tk.LogIt(tk.LogTrace, "api: IPv4 Neighbor   %s API called. url : %s\n", params.HTTPRequest.Method, params.HTTPRequest.URL)
	err := loxinlp.DelNeighNoHook(params.IPAddress, params.IfName)
	if err != nil {
		return &ErrorResponse{Payload: ResultErrorResponseErrorMessage(err.Error())}
	}
	return &ResultResponse{Result: "Success"}
}

func ConfigGetNeighbor(params operations.GetConfigNeighborAllParams, principal interface{}) middleware.Responder {
	tk.LogIt(tk.LogTrace, "api: IPv4 Neighbor  %s API called. url : %s\n", params.HTTPRequest.Method, params.HTTPRequest.URL)
	res, _ := ApiHooks.NetNeighGet()
	var result []*models.NeighborEntry
	result = make([]*models.NeighborEntry, 0)
	for _, neighbor := range res {
		var tmpResult models.NeighborEntry
		macStr := neighbor.HardwareAddr.String()
		tmpResult.MacAddress = &macStr
		ipStr := neighbor.IP.String()
		tmpResult.IPAddress = &ipStr
		devStr, _ := loxinlp.GetLinkNameByIndex(neighbor.LinkIndex)
		tmpResult.Dev = &devStr
		result = append(result, &tmpResult)
	}
	return operations.NewGetConfigNeighborAllOK().WithPayload(&operations.GetConfigNeighborAllOKBody{NeighborAttr: result})
}
