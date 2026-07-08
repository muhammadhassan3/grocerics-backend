// Package dto contains request and response data transfer objects.
package dto

import "encoding/json"

// @Swagger:model Response
// @Description: Generic response structure for API responses
// @Property status: The status of the response (e.g., "success", "failed")
// @Property message: A message providing additional information about the response
// @Property data: The actual data being returned in the response (optional)
type Response struct {
	Status  string      `json:"status"`
	Code    string      `json:"code,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message"`
}

func (data Response) ToMap() map[string]interface{} {
	var result map[string]interface{}

	jsonResult, _ := json.Marshal(data)

	json.Unmarshal(jsonResult, &result)

	return result
}
