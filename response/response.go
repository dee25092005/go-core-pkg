package response

type SuccessResponse struct {
	Data    interface{} `json:"data"`
	Message string      `json:"message,omitempty"`
}
