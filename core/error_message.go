package core

type ErrorMessage struct {
	Message string `json:"message"`
}

type ErrorMessageDTO struct {
	HttpStatusCode int            `json:"httpStatusCode"`
	ErrorMessages  []ErrorMessage `json:"errorMessages"`
}

func NewErrorMessage(err ...error) []ErrorMessage {
	errMessages := make([]ErrorMessage, len(err))
	for i, e := range err {
		errMessages[i] = ErrorMessage{Message: e.Error()}
	}

	return errMessages
}

func NewErrorMessageDTO(httpStatusCode int, err ...error) ErrorMessageDTO {
	return ErrorMessageDTO{
		HttpStatusCode: httpStatusCode,
		ErrorMessages:  NewErrorMessage(err...),
	}
}

func NewErrorMessageStr(err ...string) []ErrorMessage {
	errMessages := make([]ErrorMessage, len(err))
	for i, e := range err {
		errMessages[i] = ErrorMessage{Message: e}
	}

	return errMessages
}

func NewErrorMessageDTOStr(httpStatusCode int, message ...string) ErrorMessageDTO {
	return ErrorMessageDTO{
		HttpStatusCode: httpStatusCode,
		ErrorMessages:  NewErrorMessageStr(message...),
	}
}
