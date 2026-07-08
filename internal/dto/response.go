// Package dto contains request and response data transfer objects.
package dto

import "encoding/json"

// @Swagger:model Response
// @Property status: Outcome of the request, e.g. "success" or "failed"
// @Property code: Machine-readable error/status code, present on failures
// @Property data: Response payload, shape depends on the endpoint
// @Property message: Human-readable message describing the result
// @Description Generic response envelope wrapping every API response.
type Response struct {
	// Outcome of the request, e.g. "success" or "failed"
	Status string `json:"status"`
	// Machine-readable error/status code, present on failures
	Code string `json:"code,omitempty"`
	// Response payload, shape depends on the endpoint
	Data interface{} `json:"data,omitempty"`
	// Human-readable message describing the result
	Message string `json:"message"`
}

func (data Response) ToMap() map[string]interface{} {
	var result map[string]interface{}

	jsonResult, _ := json.Marshal(data)

	json.Unmarshal(jsonResult, &result)

	return result
}
