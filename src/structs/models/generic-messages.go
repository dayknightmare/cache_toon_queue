package models

type ErrorMessage struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type ResponseError struct {
	Message interface{} `json:"message"`
}

type ResponseSuccess struct {
	Data interface{} `json:"data"`
}
