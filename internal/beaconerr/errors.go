package beaconerr

import (
	"errors"
)

type ErrorCode string

const (
	ErrorCodeZoneAlreadyExists           ErrorCode = "ZoneAlreadyExists"
	ErrorCodeNoSuchZone                  ErrorCode = "NoSuchZone"
	ErrorCodeResponsePolicyAlreadyExists ErrorCode = "ResponsePolicyAlreadyExists"
	ErrorCodeNoSuchResponsePolicy        ErrorCode = "NoSuchResponsePolicy"
	ErrorCodeNoSuchResponsePolicyRule    ErrorCode = "NoSuchResponsePolicyRule"
	ErrorCodeNoSuchChange                ErrorCode = "NoSuchChange"
	ErrorCodeNoSuchResourceRecordSet     ErrorCode = "NoSuchResourceRecordSet"
	ErrorCodeInvalidArgument             ErrorCode = "InvalidArgument"
	ErrorCodeInternalError               ErrorCode = "InternalError"
)

func (e ErrorCode) String() string {
	return string(e)
}

type BeaconError struct {
	code    ErrorCode
	message *string
	cause   error
}

var _ error = (*BeaconError)(nil)

func NewBeaconError(code ErrorCode, message string, cause error) *BeaconError {
	return &BeaconError{
		code:    code,
		message: &message,
		cause:   cause,
	}
}

func (e *BeaconError) Code() string {
	return string(e.code)
}

func (e *BeaconError) Message() string {
	if e.message != nil {
		return *e.message
	}
	return ""
}

func (e *BeaconError) Cause() error {
	return e.cause
}

func (e *BeaconError) Error() string {
	msg := e.Code()

	if e.Message() != "" {
		msg += ": " + e.Message()
	}

	if e.cause != nil {
		msg += ": " + e.cause.Error()
	}

	return msg
}

func (e *BeaconError) Unwrap() error {
	return e.cause
}

type NoSuchError struct {
	*BeaconError
}

func (e *NoSuchError) Unwrap() error {
	return e.BeaconError
}

func newNoSuchError(code ErrorCode, message string) *NoSuchError {
	return &NoSuchError{
		BeaconError: &BeaconError{
			code:    code,
			message: &message,
			cause:   nil,
		},
	}
}

type InternalError struct {
	*BeaconError
}

func (e *InternalError) Unwrap() error {
	return e.BeaconError
}

type ConflictError struct {
	*BeaconError
}

func (e *ConflictError) Unwrap() error {
	return e.BeaconError
}

func newConflictError(code ErrorCode, message string) *ConflictError {
	return &ConflictError{
		BeaconError: &BeaconError{
			code:    code,
			message: &message,
			cause:   nil,
		},
	}
}

type BadRequestError struct {
	*BeaconError
}

func (e *BadRequestError) Unwrap() error {
	return e.BeaconError
}

func newBadRequestError(code ErrorCode, message string, cause error) *BadRequestError {
	return &BadRequestError{
		BeaconError: &BeaconError{
			code:    code,
			message: &message,
			cause:   cause,
		},
	}
}

type ZoneAlreadyExistsError struct {
	*ConflictError
}

func (e *ZoneAlreadyExistsError) Unwrap() error {
	return e.ConflictError
}

func ErrZoneAlreadyExists(message string) *ZoneAlreadyExistsError {
	return &ZoneAlreadyExistsError{
		ConflictError: newConflictError(ErrorCodeZoneAlreadyExists, message),
	}
}

type NoSuchZoneError struct {
	*NoSuchError
}

func (e *NoSuchZoneError) Unwrap() error {
	return e.NoSuchError
}

func ErrNoSuchZone(message string) *NoSuchZoneError {
	return &NoSuchZoneError{
		NoSuchError: newNoSuchError(ErrorCodeNoSuchZone, message),
	}
}

type NoSuchChangeError struct {
	*NoSuchError
}

func (e *NoSuchChangeError) Unwrap() error {
	return e.NoSuchError
}

func ErrNoSuchChange(message string) *NoSuchChangeError {
	return &NoSuchChangeError{
		NoSuchError: newNoSuchError(ErrorCodeNoSuchChange, message),
	}
}

type NoSuchResourceRecordSetError struct {
	*NoSuchError
}

func (e *NoSuchResourceRecordSetError) Unwrap() error {
	return e.NoSuchError
}

func ErrNoSuchResourceRecordSet(message string) *NoSuchResourceRecordSetError {
	return &NoSuchResourceRecordSetError{
		NoSuchError: newNoSuchError(ErrorCodeNoSuchResourceRecordSet, message),
	}
}

type ResponsePolicyAlreadyExistsError struct {
	*ConflictError
}

func (e *ResponsePolicyAlreadyExistsError) Unwrap() error {
	return e.ConflictError
}

func ErrResponsePolicyAlreadyExists(message string) *ResponsePolicyAlreadyExistsError {
	return &ResponsePolicyAlreadyExistsError{
		ConflictError: newConflictError(ErrorCodeResponsePolicyAlreadyExists, message),
	}
}

type NoSuchResponsePolicyRuleError struct {
	*NoSuchError
}

func (e *NoSuchResponsePolicyRuleError) Unwrap() error {
	return e.NoSuchError
}

func ErrNoSuchResponsePolicyRule(message string) *NoSuchResponsePolicyRuleError {
	return &NoSuchResponsePolicyRuleError{
		NoSuchError: newNoSuchError(ErrorCodeNoSuchResponsePolicyRule, message),
	}
}

type NoSuchResponsePolicyError struct {
	*NoSuchError
}

func (e *NoSuchResponsePolicyError) Unwrap() error {
	return e.NoSuchError
}

func ErrNoSuchResponsePolicy(message string) *NoSuchResponsePolicyError {
	return &NoSuchResponsePolicyError{
		NoSuchError: newNoSuchError(ErrorCodeNoSuchResponsePolicy, message),
	}
}

func ErrInternalError(message string, cause error) *InternalError {
	return &InternalError{
		BeaconError: NewBeaconError(ErrorCodeInternalError, message, cause),
	}
}

type InvalidArgumentError struct {
	*BadRequestError
	Argument string
}

func (e *InvalidArgumentError) Unwrap() error {
	return e.BadRequestError
}

func ErrInvalidArgument(message string, argument string) *InvalidArgumentError {
	return &InvalidArgumentError{
		BadRequestError: newBadRequestError(ErrorCodeInvalidArgument, message, nil),
		Argument:        argument,
	}
}

func IsNoSuchError(err error) bool {
	var noSuchErr *NoSuchError
	return errors.As(err, &noSuchErr)
}

func IsConflictError(err error) bool {
	var conflictErr *ConflictError
	return errors.As(err, &conflictErr)
}

func IsInternalError(err error) bool {
	var internalErr *InternalError
	return errors.As(err, &internalErr)
}

func IsBadRequestError(err error) bool {
	var badRequestErr *BadRequestError
	return errors.As(err, &badRequestErr)
}
