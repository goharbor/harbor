package errors

const (
	// NotFoundCode is code for the error of no object found
	NotFoundCode = "NOT_FOUND"
	// ConflictCode ...
	ConflictCode = "CONFLICT"
	// UnAuthorizedCode ...
	UnAuthorizedCode = "UNAUTHORIZED"
	// BadRequestCode ...
	BadRequestCode = "BAD_REQUEST"
	// ForbiddenCode ...
	ForbiddenCode = "FORBIDDEN"
	// PreconditionCode ...
	PreconditionCode = "PRECONDITION"
	// GeneralCode ...
	GeneralCode = "UNKNOWN"
	// DENIED it's used by middleware(readonly, vul and content trust) and returned to docker client to index the request is denied.
	DENIED = "DENIED"
	// PROJECTPOLICYVIOLATION ...
	PROJECTPOLICYVIOLATION = "PROJECTPOLICYVIOLATION"
	// ViolateForeignKeyConstraintCode is the error code for violating foreign key constraint error
	ViolateForeignKeyConstraintCode = "VIOLATE_FOREIGN_KEY_CONSTRAINT"
	// DIGESTINVALID ...
	DIGESTINVALID = "DIGEST_INVALID"
)

// NotFoundError is error for the case of object not found
func NotFoundError(err error) *Error {
	return New(err).WithCode(NotFoundCode).WithMessage("resource not found")
}

// ConflictError is error for the case of object conflict
func ConflictError(err error) *Error {
	return New(err).WithCode(ConflictCode).WithMessage("resource conflict")
}

// DeniedError is error for the case of denied
func DeniedError(err error) *Error {
	return New(err).WithCode(DENIED).WithMessage("denied")
}

// UnauthorizedError is error for the case of unauthorized accessing
func UnauthorizedError(err error) *Error {
	return New(err).WithCode(UnAuthorizedCode).WithMessage("unauthorized")
}

// BadRequestError is error for the case of bad request
func BadRequestError(err error) *Error {
	return New(err).WithCode(BadRequestCode).WithMessage("bad request")
}

// ForbiddenError is error for the case of forbidden
func ForbiddenError(err error) *Error {
	return New(err).WithCode(ForbiddenCode).WithMessage("forbidden")
}

// PreconditionFailedError is error for the case of precondition failed
func PreconditionFailedError(err error) *Error {
	return New(err).WithCode(PreconditionCode).WithMessage("precondition failed")
}

// UnknownError ...
func UnknownError(err error) *Error {
	return New(err).WithCode(GeneralCode).WithMessage("unknown")
}
