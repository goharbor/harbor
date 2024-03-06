// Package pgerrcode contains constants for PostgreSQL error codes.
package pgerrcode

// Source: https://www.postgresql.org/docs/14/errcodes-appendix.html
// See gen.rb for script that can convert the error code table to Go code.

const (

	// Class 00 — Successful Completion
	SuccessfulCompletion = "00000"

	// Class 01 — Warning
	Warning                          = "01000"
	DynamicResultSetsReturned        = "0100C"
	ImplicitZeroBitPadding           = "01008"
	NullValueEliminatedInSetFunction = "01003"
	PrivilegeNotGranted              = "01007"
	PrivilegeNotRevoked              = "01006"
	StringDataRightTruncationWarning = "01004"
	DeprecatedFeature                = "01P01"

	// Class 02 — No Data (this is also a warning class per the SQL standard)
	NoData                                = "02000"
	NoAdditionalDynamicResultSetsReturned = "02001"

	// Class 03 — SQL Statement Not Yet Complete
	SQLStatementNotYetComplete = "03000"

	// Class 08 — Connection Exception
	ConnectionException                           = "08000"
	ConnectionDoesNotExist                        = "08003"
	ConnectionFailure                             = "08006"
	SQLClientUnableToEstablishSQLConnection       = "08001"
	SQLServerRejectedEstablishmentOfSQLConnection = "08004"
	TransactionResolutionUnknown                  = "08007"
	ProtocolViolation                             = "08P01"

	// Class 09 — Triggered Action Exception
	TriggeredActionException = "09000"

	// Class 0A — Feature Not Supported
	FeatureNotSupported = "0A000"

	// Class 0B — Invalid Transaction Initiation
	InvalidTransactionInitiation = "0B000"

	// Class 0F — Locator Exception
	LocatorException            = "0F000"
	InvalidLocatorSpecification = "0F001"

	// Class 0L — Invalid Grantor
	InvalidGrantor        = "0L000"
	InvalidGrantOperation = "0LP01"

	// Class 0P — Invalid Role Specification
	InvalidRoleSpecification = "0P000"

	// Class 0Z — Diagnostics Exception
	DiagnosticsException                           = "0Z000"
	StackedDiagnosticsAccessedWithoutActiveHandler = "0Z002"

	// Class 20 — Case Not Found
	CaseNotFound = "20000"

	// Class 21 — Cardinality Violation
	CardinalityViolation = "21000"

	// Class 22 — Data Exception
	DataException                             = "22000"
	ArraySubscriptError                       = "2202E"
	CharacterNotInRepertoire                  = "22021"
	DatetimeFieldOverflow                     = "22008"
	DivisionByZero                            = "22012"
	ErrorInAssignment                         = "22005"
	EscapeCharacterConflict                   = "2200B"
	IndicatorOverflow                         = "22022"
	IntervalFieldOverflow                     = "22015"
	InvalidArgumentForLogarithm               = "2201E"
	InvalidArgumentForNtileFunction           = "22014"
	InvalidArgumentForNthValueFunction        = "22016"
	InvalidArgumentForPowerFunction           = "2201F"
	InvalidArgumentForWidthBucketFunction     = "2201G"
	InvalidCharacterValueForCast              = "22018"
	InvalidDatetimeFormat                     = "22007"
	InvalidEscapeCharacter                    = "22019"
	InvalidEscapeOctet                        = "2200D"
	InvalidEscapeSequence                     = "22025"
	NonstandardUseOfEscapeCharacter           = "22P06"
	InvalidIndicatorParameterValue            = "22010"
	InvalidParameterValue                     = "22023"
	InvalidPrecedingOrFollowingSize           = "22013"
	InvalidRegularExpression                  = "2201B"
	InvalidRowCountInLimitClause              = "2201W"
	InvalidRowCountInResultOffsetClause       = "2201X"
	InvalidTablesampleArgument                = "2202H"
	InvalidTablesampleRepeat                  = "2202G"
	InvalidTimeZoneDisplacementValue          = "22009"
	InvalidUseOfEscapeCharacter               = "2200C"
	MostSpecificTypeMismatch                  = "2200G"
	NullValueNotAllowedDataException          = "22004"
	NullValueNoIndicatorParameter             = "22002"
	NumericValueOutOfRange                    = "22003"
	SequenceGeneratorLimitExceeded            = "2200H"
	StringDataLengthMismatch                  = "22026"
	StringDataRightTruncationDataException    = "22001"
	SubstringError                            = "22011"
	TrimError                                 = "22027"
	UnterminatedCString                       = "22024"
	ZeroLengthCharacterString                 = "2200F"
	FloatingPointException                    = "22P01"
	InvalidTextRepresentation                 = "22P02"
	InvalidBinaryRepresentation               = "22P03"
	BadCopyFileFormat                         = "22P04"
	UntranslatableCharacter                   = "22P05"
	NotAnXMLDocument                          = "2200L"
	InvalidXMLDocument                        = "2200M"
	InvalidXMLContent                         = "2200N"
	InvalidXMLComment                         = "2200S"
	InvalidXMLProcessingInstruction           = "2200T"
	DuplicateJSONObjectKeyValue               = "22030"
	InvalidArgumentForSQLJSONDatetimeFunction = "22031"
	InvalidJSONText                           = "22032"
	InvalidSQLJSONSubscript                   = "22033"
	MoreThanOneSQLJSONItem                    = "22034"
	NoSQLJSONItem                             = "22035"
	NonNumericSQLJSONItem                     = "22036"
	NonUniqueKeysInAJSONObject                = "22037"
	SingletonSQLJSONItemRequired              = "22038"
	SQLJSONArrayNotFound                      = "22039"
	SQLJSONMemberNotFound                     = "2203A"
	SQLJSONNumberNotFound                     = "2203B"
	SQLJSONObjectNotFound                     = "2203C"
	TooManyJSONArrayElements                  = "2203D"
	TooManyJSONObjectMembers                  = "2203E"
	SQLJSONScalarRequired                     = "2203F"

	// Class 23 — Integrity Constraint Violation
	IntegrityConstraintViolation = "23000"
	RestrictViolation            = "23001"
	NotNullViolation             = "23502"
	ForeignKeyViolation          = "23503"
	UniqueViolation              = "23505"
	CheckViolation               = "23514"
	ExclusionViolation           = "23P01"

	// Class 24 — Invalid Cursor State
	InvalidCursorState = "24000"

	// Class 25 — Invalid Transaction State
	InvalidTransactionState                         = "25000"
	ActiveSQLTransaction                            = "25001"
	BranchTransactionAlreadyActive                  = "25002"
	HeldCursorRequiresSameIsolationLevel            = "25008"
	InappropriateAccessModeForBranchTransaction     = "25003"
	InappropriateIsolationLevelForBranchTransaction = "25004"
	NoActiveSQLTransactionForBranchTransaction      = "25005"
	ReadOnlySQLTransaction                          = "25006"
	SchemaAndDataStatementMixingNotSupported        = "25007"
	NoActiveSQLTransaction                          = "25P01"
	InFailedSQLTransaction                          = "25P02"
	IdleInTransactionSessionTimeout                 = "25P03"

	// Class 26 — Invalid SQL Statement Name
	InvalidSQLStatementName = "26000"

	// Class 27 — Triggered Data Change Violation
	TriggeredDataChangeViolation = "27000"

	// Class 28 — Invalid Authorization Specification
	InvalidAuthorizationSpecification = "28000"
	InvalidPassword                   = "28P01"

	// Class 2B — Dependent Privilege Descriptors Still Exist
	DependentPrivilegeDescriptorsStillExist = "2B000"
	DependentObjectsStillExist              = "2BP01"

	// Class 2D — Invalid Transaction Termination
	InvalidTransactionTermination = "2D000"

	// Class 2F — SQL Routine Exception
	SQLRoutineException                                = "2F000"
	FunctionExecutedNoReturnStatement                  = "2F005"
	ModifyingSQLDataNotPermittedSQLRoutineException    = "2F002"
	ProhibitedSQLStatementAttemptedSQLRoutineException = "2F003"
	ReadingSQLDataNotPermittedSQLRoutineException      = "2F004"

	// Class 34 — Invalid Cursor Name
	InvalidCursorName = "34000"

	// Class 38 — External Routine Exception
	ExternalRoutineException                                = "38000"
	ContainingSQLNotPermitted                               = "38001"
	ModifyingSQLDataNotPermittedExternalRoutineException    = "38002"
	ProhibitedSQLStatementAttemptedExternalRoutineException = "38003"
	ReadingSQLDataNotPermittedExternalRoutineException      = "38004"

	// Class 39 — External Routine Invocation Exception
	ExternalRoutineInvocationException                    = "39000"
	InvalidSQLstateReturned                               = "39001"
	NullValueNotAllowedExternalRoutineInvocationException = "39004"
	TriggerProtocolViolated                               = "39P01"
	SRFProtocolViolated                                   = "39P02"
	EventTriggerProtocolViolated                          = "39P03"

	// Class 3B — Savepoint Exception
	SavepointException            = "3B000"
	InvalidSavepointSpecification = "3B001"

	// Class 3D — Invalid Catalog Name
	InvalidCatalogName = "3D000"

	// Class 3F — Invalid Schema Name
	InvalidSchemaName = "3F000"

	// Class 40 — Transaction Rollback
	TransactionRollback                     = "40000"
	TransactionIntegrityConstraintViolation = "40002"
	SerializationFailure                    = "40001"
	StatementCompletionUnknown              = "40003"
	DeadlockDetected                        = "40P01"

	// Class 42 — Syntax Error or Access Rule Violation
	SyntaxErrorOrAccessRuleViolation   = "42000"
	SyntaxError                        = "42601"
	InsufficientPrivilege              = "42501"
	CannotCoerce                       = "42846"
	GroupingError                      = "42803"
	WindowingError                     = "42P20"
	InvalidRecursion                   = "42P19"
	InvalidForeignKey                  = "42830"
	InvalidName                        = "42602"
	NameTooLong                        = "42622"
	ReservedName                       = "42939"
	DatatypeMismatch                   = "42804"
	IndeterminateDatatype              = "42P18"
	CollationMismatch                  = "42P21"
	IndeterminateCollation             = "42P22"
	WrongObjectType                    = "42809"
	GeneratedAlways                    = "428C9"
	UndefinedColumn                    = "42703"
	UndefinedFunction                  = "42883"
	UndefinedTable                     = "42P01"
	UndefinedParameter                 = "42P02"
	UndefinedObject                    = "42704"
	DuplicateColumn                    = "42701"
	DuplicateCursor                    = "42P03"
	DuplicateDatabase                  = "42P04"
	DuplicateFunction                  = "42723"
	DuplicatePreparedStatement         = "42P05"
	DuplicateSchema                    = "42P06"
	DuplicateTable                     = "42P07"
	DuplicateAlias                     = "42712"
	DuplicateObject                    = "42710"
	AmbiguousColumn                    = "42702"
	AmbiguousFunction                  = "42725"
	AmbiguousParameter                 = "42P08"
	AmbiguousAlias                     = "42P09"
	InvalidColumnReference             = "42P10"
	InvalidColumnDefinition            = "42611"
	InvalidCursorDefinition            = "42P11"
	InvalidDatabaseDefinition          = "42P12"
	InvalidFunctionDefinition          = "42P13"
	InvalidPreparedStatementDefinition = "42P14"
	InvalidSchemaDefinition            = "42P15"
	InvalidTableDefinition             = "42P16"
	InvalidObjectDefinition            = "42P17"

	// Class 44 — WITH CHECK OPTION Violation
	WithCheckOptionViolation = "44000"

	// Class 53 — Insufficient Resources
	InsufficientResources      = "53000"
	DiskFull                   = "53100"
	OutOfMemory                = "53200"
	TooManyConnections         = "53300"
	ConfigurationLimitExceeded = "53400"

	// Class 54 — Program Limit Exceeded
	ProgramLimitExceeded = "54000"
	StatementTooComplex  = "54001"
	TooManyColumns       = "54011"
	TooManyArguments     = "54023"

	// Class 55 — Object Not In Prerequisite State
	ObjectNotInPrerequisiteState = "55000"
	ObjectInUse                  = "55006"
	CantChangeRuntimeParam       = "55P02"
	LockNotAvailable             = "55P03"
	UnsafeNewEnumValueUsage      = "55P04"

	// Class 57 — Operator Intervention
	OperatorIntervention = "57000"
	QueryCanceled        = "57014"
	AdminShutdown        = "57P01"
	CrashShutdown        = "57P02"
	CannotConnectNow     = "57P03"
	DatabaseDropped      = "57P04"
	IdleSessionTimeout   = "57P05"

	// Class 58 — System Error (errors external to PostgreSQL itself)
	SystemError   = "58000"
	IOError       = "58030"
	UndefinedFile = "58P01"
	DuplicateFile = "58P02"

	// Class 72 — Snapshot Failure
	SnapshotTooOld = "72000"

	// Class F0 — Configuration File Error
	ConfigFileError = "F0000"
	LockFileExists  = "F0001"

	// Class HV — Foreign Data Wrapper Error (SQL/MED)
	FDWError                             = "HV000"
	FDWColumnNameNotFound                = "HV005"
	FDWDynamicParameterValueNeeded       = "HV002"
	FDWFunctionSequenceError             = "HV010"
	FDWInconsistentDescriptorInformation = "HV021"
	FDWInvalidAttributeValue             = "HV024"
	FDWInvalidColumnName                 = "HV007"
	FDWInvalidColumnNumber               = "HV008"
	FDWInvalidDataType                   = "HV004"
	FDWInvalidDataTypeDescriptors        = "HV006"
	FDWInvalidDescriptorFieldIdentifier  = "HV091"
	FDWInvalidHandle                     = "HV00B"
	FDWInvalidOptionIndex                = "HV00C"
	FDWInvalidOptionName                 = "HV00D"
	FDWInvalidStringLengthOrBufferLength = "HV090"
	FDWInvalidStringFormat               = "HV00A"
	FDWInvalidUseOfNullPointer           = "HV009"
	FDWTooManyHandles                    = "HV014"
	FDWOutOfMemory                       = "HV001"
	FDWNoSchemas                         = "HV00P"
	FDWOptionNameNotFound                = "HV00J"
	FDWReplyHandle                       = "HV00K"
	FDWSchemaNotFound                    = "HV00Q"
	FDWTableNotFound                     = "HV00R"
	FDWUnableToCreateExecution           = "HV00L"
	FDWUnableToCreateReply               = "HV00M"
	FDWUnableToEstablishConnection       = "HV00N"

	// Class P0 — PL/pgSQL Error
	PLpgSQLError   = "P0000"
	RaiseException = "P0001"
	NoDataFound    = "P0002"
	TooManyRows    = "P0003"
	AssertFailure  = "P0004"

	// Class XX — Internal Error
	InternalError  = "XX000"
	DataCorrupted  = "XX001"
	IndexCorrupted = "XX002"
)

// IsSuccessfulCompletion asserts the error code class is Class 00 — Successful Completion
func IsSuccessfulCompletion(code string) bool {
	switch code {
	case SuccessfulCompletion:
		return true
	}
	return false
}

// IsWarning asserts the error code class is Class 01 — Warning
func IsWarning(code string) bool {
	switch code {
	case Warning, DynamicResultSetsReturned, ImplicitZeroBitPadding, NullValueEliminatedInSetFunction, PrivilegeNotGranted, PrivilegeNotRevoked, StringDataRightTruncationWarning, DeprecatedFeature:
		return true
	}
	return false
}

// IsNoData asserts the error code class is Class 02 — No Data (this is also a warning class per the SQL standard)
func IsNoData(code string) bool {
	switch code {
	case NoData, NoAdditionalDynamicResultSetsReturned:
		return true
	}
	return false
}

// IsSQLStatementNotYetComplete asserts the error code class is Class 03 — SQL Statement Not Yet Complete
func IsSQLStatementNotYetComplete(code string) bool {
	switch code {
	case SQLStatementNotYetComplete:
		return true
	}
	return false
}

// IsConnectionException asserts the error code class is Class 08 — Connection Exception
func IsConnectionException(code string) bool {
	switch code {
	case ConnectionException, ConnectionDoesNotExist, ConnectionFailure, SQLClientUnableToEstablishSQLConnection, SQLServerRejectedEstablishmentOfSQLConnection, TransactionResolutionUnknown, ProtocolViolation:
		return true
	}
	return false
}

// IsTriggeredActionException asserts the error code class is Class 09 — Triggered Action Exception
func IsTriggeredActionException(code string) bool {
	switch code {
	case TriggeredActionException:
		return true
	}
	return false
}

// IsFeatureNotSupported asserts the error code class is Class 0A — Feature Not Supported
func IsFeatureNotSupported(code string) bool {
	switch code {
	case FeatureNotSupported:
		return true
	}
	return false
}

// IsInvalidTransactionInitiation asserts the error code class is Class 0B — Invalid Transaction Initiation
func IsInvalidTransactionInitiation(code string) bool {
	switch code {
	case InvalidTransactionInitiation:
		return true
	}
	return false
}

// IsLocatorException asserts the error code class is Class 0F — Locator Exception
func IsLocatorException(code string) bool {
	switch code {
	case LocatorException, InvalidLocatorSpecification:
		return true
	}
	return false
}

// IsInvalidGrantor asserts the error code class is Class 0L — Invalid Grantor
func IsInvalidGrantor(code string) bool {
	switch code {
	case InvalidGrantor, InvalidGrantOperation:
		return true
	}
	return false
}

// IsInvalidRoleSpecification asserts the error code class is Class 0P — Invalid Role Specification
func IsInvalidRoleSpecification(code string) bool {
	switch code {
	case InvalidRoleSpecification:
		return true
	}
	return false
}

// IsDiagnosticsException asserts the error code class is Class 0Z — Diagnostics Exception
func IsDiagnosticsException(code string) bool {
	switch code {
	case DiagnosticsException, StackedDiagnosticsAccessedWithoutActiveHandler:
		return true
	}
	return false
}

// IsCaseNotFound asserts the error code class is Class 20 — Case Not Found
func IsCaseNotFound(code string) bool {
	switch code {
	case CaseNotFound:
		return true
	}
	return false
}

// IsCardinalityViolation asserts the error code class is Class 21 — Cardinality Violation
func IsCardinalityViolation(code string) bool {
	switch code {
	case CardinalityViolation:
		return true
	}
	return false
}

// IsDataException asserts the error code class is Class 22 — Data Exception
func IsDataException(code string) bool {
	switch code {
	case DataException, ArraySubscriptError, CharacterNotInRepertoire, DatetimeFieldOverflow, DivisionByZero, ErrorInAssignment, EscapeCharacterConflict, IndicatorOverflow, IntervalFieldOverflow, InvalidArgumentForLogarithm, InvalidArgumentForNtileFunction, InvalidArgumentForNthValueFunction, InvalidArgumentForPowerFunction, InvalidArgumentForWidthBucketFunction, InvalidCharacterValueForCast, InvalidDatetimeFormat, InvalidEscapeCharacter, InvalidEscapeOctet, InvalidEscapeSequence, NonstandardUseOfEscapeCharacter, InvalidIndicatorParameterValue, InvalidParameterValue, InvalidPrecedingOrFollowingSize, InvalidRegularExpression, InvalidRowCountInLimitClause, InvalidRowCountInResultOffsetClause, InvalidTablesampleArgument, InvalidTablesampleRepeat, InvalidTimeZoneDisplacementValue, InvalidUseOfEscapeCharacter, MostSpecificTypeMismatch, NullValueNotAllowedDataException, NullValueNoIndicatorParameter, NumericValueOutOfRange, SequenceGeneratorLimitExceeded, StringDataLengthMismatch, StringDataRightTruncationDataException, SubstringError, TrimError, UnterminatedCString, ZeroLengthCharacterString, FloatingPointException, InvalidTextRepresentation, InvalidBinaryRepresentation, BadCopyFileFormat, UntranslatableCharacter, NotAnXMLDocument, InvalidXMLDocument, InvalidXMLContent, InvalidXMLComment, InvalidXMLProcessingInstruction, DuplicateJSONObjectKeyValue, InvalidArgumentForSQLJSONDatetimeFunction, InvalidJSONText, InvalidSQLJSONSubscript, MoreThanOneSQLJSONItem, NoSQLJSONItem, NonNumericSQLJSONItem, NonUniqueKeysInAJSONObject, SingletonSQLJSONItemRequired, SQLJSONArrayNotFound, SQLJSONMemberNotFound, SQLJSONNumberNotFound, SQLJSONObjectNotFound, TooManyJSONArrayElements, TooManyJSONObjectMembers, SQLJSONScalarRequired:
		return true
	}
	return false
}

// IsIntegrityConstraintViolation asserts the error code class is Class 23 — Integrity Constraint Violation
func IsIntegrityConstraintViolation(code string) bool {
	switch code {
	case IntegrityConstraintViolation, RestrictViolation, NotNullViolation, ForeignKeyViolation, UniqueViolation, CheckViolation, ExclusionViolation:
		return true
	}
	return false
}

// IsInvalidCursorState asserts the error code class is Class 24 — Invalid Cursor State
func IsInvalidCursorState(code string) bool {
	switch code {
	case InvalidCursorState:
		return true
	}
	return false
}

// IsInvalidTransactionState asserts the error code class is Class 25 — Invalid Transaction State
func IsInvalidTransactionState(code string) bool {
	switch code {
	case InvalidTransactionState, ActiveSQLTransaction, BranchTransactionAlreadyActive, HeldCursorRequiresSameIsolationLevel, InappropriateAccessModeForBranchTransaction, InappropriateIsolationLevelForBranchTransaction, NoActiveSQLTransactionForBranchTransaction, ReadOnlySQLTransaction, SchemaAndDataStatementMixingNotSupported, NoActiveSQLTransaction, InFailedSQLTransaction, IdleInTransactionSessionTimeout:
		return true
	}
	return false
}

// IsInvalidSQLStatementName asserts the error code class is Class 26 — Invalid SQL Statement Name
func IsInvalidSQLStatementName(code string) bool {
	switch code {
	case InvalidSQLStatementName:
		return true
	}
	return false
}

// IsTriggeredDataChangeViolation asserts the error code class is Class 27 — Triggered Data Change Violation
func IsTriggeredDataChangeViolation(code string) bool {
	switch code {
	case TriggeredDataChangeViolation:
		return true
	}
	return false
}

// IsInvalidAuthorizationSpecification asserts the error code class is Class 28 — Invalid Authorization Specification
func IsInvalidAuthorizationSpecification(code string) bool {
	switch code {
	case InvalidAuthorizationSpecification, InvalidPassword:
		return true
	}
	return false
}

// IsDependentPrivilegeDescriptorsStillExist asserts the error code class is Class 2B — Dependent Privilege Descriptors Still Exist
func IsDependentPrivilegeDescriptorsStillExist(code string) bool {
	switch code {
	case DependentPrivilegeDescriptorsStillExist, DependentObjectsStillExist:
		return true
	}
	return false
}

// IsInvalidTransactionTermination asserts the error code class is Class 2D — Invalid Transaction Termination
func IsInvalidTransactionTermination(code string) bool {
	switch code {
	case InvalidTransactionTermination:
		return true
	}
	return false
}

// IsSQLRoutineException asserts the error code class is Class 2F — SQL Routine Exception
func IsSQLRoutineException(code string) bool {
	switch code {
	case SQLRoutineException, FunctionExecutedNoReturnStatement, ModifyingSQLDataNotPermittedSQLRoutineException, ProhibitedSQLStatementAttemptedSQLRoutineException, ReadingSQLDataNotPermittedSQLRoutineException:
		return true
	}
	return false
}

// IsInvalidCursorName asserts the error code class is Class 34 — Invalid Cursor Name
func IsInvalidCursorName(code string) bool {
	switch code {
	case InvalidCursorName:
		return true
	}
	return false
}

// IsExternalRoutineException asserts the error code class is Class 38 — External Routine Exception
func IsExternalRoutineException(code string) bool {
	switch code {
	case ExternalRoutineException, ContainingSQLNotPermitted, ModifyingSQLDataNotPermittedExternalRoutineException, ProhibitedSQLStatementAttemptedExternalRoutineException, ReadingSQLDataNotPermittedExternalRoutineException:
		return true
	}
	return false
}

// IsExternalRoutineInvocationException asserts the error code class is Class 39 — External Routine Invocation Exception
func IsExternalRoutineInvocationException(code string) bool {
	switch code {
	case ExternalRoutineInvocationException, InvalidSQLstateReturned, NullValueNotAllowedExternalRoutineInvocationException, TriggerProtocolViolated, SRFProtocolViolated, EventTriggerProtocolViolated:
		return true
	}
	return false
}

// IsSavepointException asserts the error code class is Class 3B — Savepoint Exception
func IsSavepointException(code string) bool {
	switch code {
	case SavepointException, InvalidSavepointSpecification:
		return true
	}
	return false
}

// IsInvalidCatalogName asserts the error code class is Class 3D — Invalid Catalog Name
func IsInvalidCatalogName(code string) bool {
	switch code {
	case InvalidCatalogName:
		return true
	}
	return false
}

// IsInvalidSchemaName asserts the error code class is Class 3F — Invalid Schema Name
func IsInvalidSchemaName(code string) bool {
	switch code {
	case InvalidSchemaName:
		return true
	}
	return false
}

// IsTransactionRollback asserts the error code class is Class 40 — Transaction Rollback
func IsTransactionRollback(code string) bool {
	switch code {
	case TransactionRollback, TransactionIntegrityConstraintViolation, SerializationFailure, StatementCompletionUnknown, DeadlockDetected:
		return true
	}
	return false
}

// IsSyntaxErrororAccessRuleViolation asserts the error code class is Class 42 — Syntax Error or Access Rule Violation
func IsSyntaxErrororAccessRuleViolation(code string) bool {
	switch code {
	case SyntaxErrorOrAccessRuleViolation, SyntaxError, InsufficientPrivilege, CannotCoerce, GroupingError, WindowingError, InvalidRecursion, InvalidForeignKey, InvalidName, NameTooLong, ReservedName, DatatypeMismatch, IndeterminateDatatype, CollationMismatch, IndeterminateCollation, WrongObjectType, GeneratedAlways, UndefinedColumn, UndefinedFunction, UndefinedTable, UndefinedParameter, UndefinedObject, DuplicateColumn, DuplicateCursor, DuplicateDatabase, DuplicateFunction, DuplicatePreparedStatement, DuplicateSchema, DuplicateTable, DuplicateAlias, DuplicateObject, AmbiguousColumn, AmbiguousFunction, AmbiguousParameter, AmbiguousAlias, InvalidColumnReference, InvalidColumnDefinition, InvalidCursorDefinition, InvalidDatabaseDefinition, InvalidFunctionDefinition, InvalidPreparedStatementDefinition, InvalidSchemaDefinition, InvalidTableDefinition, InvalidObjectDefinition:
		return true
	}
	return false
}

// IsWithCheckOptionViolation asserts the error code class is Class 44 — WITH CHECK OPTION Violation
func IsWithCheckOptionViolation(code string) bool {
	switch code {
	case WithCheckOptionViolation:
		return true
	}
	return false
}

// IsInsufficientResources asserts the error code class is Class 53 — Insufficient Resources
func IsInsufficientResources(code string) bool {
	switch code {
	case InsufficientResources, DiskFull, OutOfMemory, TooManyConnections, ConfigurationLimitExceeded:
		return true
	}
	return false
}

// IsProgramLimitExceeded asserts the error code class is Class 54 — Program Limit Exceeded
func IsProgramLimitExceeded(code string) bool {
	switch code {
	case ProgramLimitExceeded, StatementTooComplex, TooManyColumns, TooManyArguments:
		return true
	}
	return false
}

// IsObjectNotInPrerequisiteState asserts the error code class is Class 55 — Object Not In Prerequisite State
func IsObjectNotInPrerequisiteState(code string) bool {
	switch code {
	case ObjectNotInPrerequisiteState, ObjectInUse, CantChangeRuntimeParam, LockNotAvailable, UnsafeNewEnumValueUsage:
		return true
	}
	return false
}

// IsOperatorIntervention asserts the error code class is Class 57 — Operator Intervention
func IsOperatorIntervention(code string) bool {
	switch code {
	case OperatorIntervention, QueryCanceled, AdminShutdown, CrashShutdown, CannotConnectNow, DatabaseDropped, IdleSessionTimeout:
		return true
	}
	return false
}

// IsSystemError asserts the error code class is Class 58 — System Error (errors external to PostgreSQL itself)
func IsSystemError(code string) bool {
	switch code {
	case SystemError, IOError, UndefinedFile, DuplicateFile:
		return true
	}
	return false
}

// IsSnapshotFailure asserts the error code class is Class 72 — Snapshot Failure
func IsSnapshotFailure(code string) bool {
	switch code {
	case SnapshotTooOld:
		return true
	}
	return false
}

// IsConfigurationFileError asserts the error code class is Class F0 — Configuration File Error
func IsConfigurationFileError(code string) bool {
	switch code {
	case ConfigFileError, LockFileExists:
		return true
	}
	return false
}

// IsForeignDataWrapperError asserts the error code class is Class HV — Foreign Data Wrapper Error (SQL/MED)
func IsForeignDataWrapperError(code string) bool {
	switch code {
	case FDWError, FDWColumnNameNotFound, FDWDynamicParameterValueNeeded, FDWFunctionSequenceError, FDWInconsistentDescriptorInformation, FDWInvalidAttributeValue, FDWInvalidColumnName, FDWInvalidColumnNumber, FDWInvalidDataType, FDWInvalidDataTypeDescriptors, FDWInvalidDescriptorFieldIdentifier, FDWInvalidHandle, FDWInvalidOptionIndex, FDWInvalidOptionName, FDWInvalidStringLengthOrBufferLength, FDWInvalidStringFormat, FDWInvalidUseOfNullPointer, FDWTooManyHandles, FDWOutOfMemory, FDWNoSchemas, FDWOptionNameNotFound, FDWReplyHandle, FDWSchemaNotFound, FDWTableNotFound, FDWUnableToCreateExecution, FDWUnableToCreateReply, FDWUnableToEstablishConnection:
		return true
	}
	return false
}

// IsPLpgSQLError asserts the error code class is Class P0 — PL/pgSQL Error
func IsPLpgSQLError(code string) bool {
	switch code {
	case PLpgSQLError, RaiseException, NoDataFound, TooManyRows, AssertFailure:
		return true
	}
	return false
}

// IsInternalError asserts the error code class is Class XX — Internal Error
func IsInternalError(code string) bool {
	switch code {
	case InternalError, DataCorrupted, IndexCorrupted:
		return true
	}
	return false
}
