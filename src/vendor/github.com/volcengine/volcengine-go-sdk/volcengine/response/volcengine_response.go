package response

type VolcengineResponse struct {
	ResponseMetadata *ResponseMetadata
	Result           interface{}
}

type ResponseMetadata struct {
	RequestId string
	Action    string
	Version   string
	Service   string
	Region    string
	HTTPCode  int
	Error     *Error
}

type Error struct {
	CodeN   int
	Code    string
	Message string
}

type VolcengineSimpleError struct {
	HttpCode  int    `json:"HTTPCode"`
	ErrorCode string `json:"errorcode"`
	Message   string `json:"message"`
}
