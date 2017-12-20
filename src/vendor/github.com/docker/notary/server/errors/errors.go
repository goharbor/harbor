package errors

import (
	"net/http"

	"github.com/docker/distribution/registry/api/errcode"
)

// The notary API is on version 1, but URLs start with /v2/ to be consistent
// with the registry API
const errGroup = "notary.api.v1"

// These errors should be returned from contextHandlers only. They are
// serialized and returned to a user as part of the generic error handling
// done by the rootHandler
var (
	ErrNoStorage = errcode.Register(errGroup, errcode.ErrorDescriptor{
		Value:          "NO_STORAGE",
		Message:        "The server is misconfigured and has no storage.",
		Description:    "No storage backend has been configured for the server.",
		HTTPStatusCode: http.StatusInternalServerError,
	})
	ErrNoFilename = errcode.Register(errGroup, errcode.ErrorDescriptor{
		Value:          "NO_FILENAME",
		Message:        "No file/role name provided.",
		Description:    "No file/role name is provided to associate an update with.",
		HTTPStatusCode: http.StatusBadRequest,
	})
	ErrInvalidRole = errcode.Register(errGroup, errcode.ErrorDescriptor{
		Value:          "INVALID_ROLE",
		Message:        "The role you are attempting to operate on is invalid.",
		Description:    "The user attempted to operate on a role that is not deemed valid.",
		HTTPStatusCode: http.StatusBadRequest,
	})
	ErrMalformedJSON = errcode.Register(errGroup, errcode.ErrorDescriptor{
		Value:          "MALFORMED_JSON",
		Message:        "JSON sent by the client could not be parsed by the server",
		Description:    "The client sent malformed JSON.",
		HTTPStatusCode: http.StatusBadRequest,
	})
	ErrUpdating = errcode.Register(errGroup, errcode.ErrorDescriptor{
		Value:          "UPDATING",
		Message:        "An error has occurred while updating the TUF repository.",
		Description:    "An error occurred when attempting to apply an update at the storage layer.",
		HTTPStatusCode: http.StatusInternalServerError,
	})
	ErrOldVersion = errcode.Register(errGroup, errcode.ErrorDescriptor{
		Value:          "VERSION",
		Message:        "A newer version of metadata is already available.",
		Description:    "A newer version of the repository's metadata is already available in storage.",
		HTTPStatusCode: http.StatusBadRequest,
	})
	ErrMetadataNotFound = errcode.Register(errGroup, errcode.ErrorDescriptor{
		Value:          "METADATA_NOT_FOUND",
		Message:        "You have requested metadata that does not exist.",
		Description:    "The user requested metadata that is not known to the server.",
		HTTPStatusCode: http.StatusNotFound,
	})
	ErrInvalidUpdate = errcode.Register(errGroup, errcode.ErrorDescriptor{
		Value:          "INVALID_UPDATE",
		Message:        "Update sent by the client is invalid.",
		Description:    "The user-uploaded TUF data has been parsed but failed validation.",
		HTTPStatusCode: http.StatusBadRequest,
	})
	ErrMalformedUpload = errcode.Register(errGroup, errcode.ErrorDescriptor{
		Value:          "MALFORMED_UPLOAD",
		Message:        "The body of your request is malformed.",
		Description:    "The user uploaded new TUF data and the server was unable to parse it as multipart/form-data.",
		HTTPStatusCode: http.StatusBadRequest,
	})
	ErrGenericNotFound = errcode.Register(errGroup, errcode.ErrorDescriptor{
		Value:          "GENERIC_NOT_FOUND",
		Message:        "You have requested a resource that does not exist.",
		Description:    "The user requested a non-specific resource that is not known to the server.",
		HTTPStatusCode: http.StatusNotFound,
	})
	ErrNoCryptoService = errcode.Register(errGroup, errcode.ErrorDescriptor{
		Value:          "NO_CRYPTOSERVICE",
		Message:        "The server does not have a signing service configured.",
		Description:    "No signing service has been configured for the server and it has been asked to perform an operation that requires either signing, or key generation.",
		HTTPStatusCode: http.StatusInternalServerError,
	})
	ErrNoKeyAlgorithm = errcode.Register(errGroup, errcode.ErrorDescriptor{
		Value:          "NO_KEYALGORITHM",
		Message:        "The server does not have a key algorithm configured.",
		Description:    "No key algorithm has been configured for the server and it has been asked to perform an operation that requires generation.",
		HTTPStatusCode: http.StatusInternalServerError,
	})
	ErrInvalidGUN = errcode.Register(errGroup, errcode.ErrorDescriptor{
		Value:          "INVALID_GUN",
		Message:        "The server does not support actions on images of this name.",
		Description:    "The server does not support actions on images of this name.",
		HTTPStatusCode: http.StatusBadRequest,
	})
	ErrInvalidParams = errcode.Register(errGroup, errcode.ErrorDescriptor{
		Value:          "INVALID_PARAMETERS",
		Message:        "The parameters provided are not valid.",
		Description:    "The parameters provided are not valid.",
		HTTPStatusCode: http.StatusBadRequest,
	})
	ErrUnknown = errcode.ErrorCodeUnknown
)
