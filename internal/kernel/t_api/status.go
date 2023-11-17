// In our system, errors are separated into two categories - platform errors and application errors.
// Platform errors represent failures at the runtime level, such as database connection issues, file I/O failures,
// or network request problems.These are usually transient issues that are recoverable if retried later.
// Application errors indicate errors code specific to our business logic and use cases. This separation allows us
// to handle the two types differently - platform errors may trigger retries with backoff, while application errors
// should report immediately to the users since these failures are not typically recoverable by simply retrying.
//
// In our Go system, platform errors are represented as typical Go `error` value returned from function calls. For example:
//
// dbResult, dbErr := database.Query("SELECT...")
//
// The dbErr would contain platform errors like connection failures. While application errors are returned in
// the response object, while the `error` return is `nil`.
package t_api

import (
	"fmt"
	"net/http"
	"strconv"

	grpcApi "github.com/resonatehq/resonate/internal/app/subsystems/api/grpc/api"
	"google.golang.org/genproto/googleapis/rpc/code"
)

// Application level status (2000-4999)

type ResponseStatus int

func (s ResponseStatus) String() string {
	return strconv.Itoa(int(s))
}

// methods to map to http status code
func (s ResponseStatus) HTTP() int {
	return int(s) / 10
}

// we capture the type of ok status in the response object to have the same dedup info as the http api
func (s ResponseStatus) GRPC_OK() grpcApi.Status {
	switch s {
	case StatusOK:
		return grpcApi.Status(http.StatusOK)
	case StatusCreated:
		return grpcApi.Status(http.StatusCreated)
	default:
		panic(fmt.Sprintf("invalid success status: %d", s))
	}
}

const (
	StatusOK                     ResponseStatus = 2000
	StatusCreated                ResponseStatus = 2010
	StatusNoContent              ResponseStatus = 2040
	StatusFieldValidationFailure ResponseStatus = 4000
	StatusPromiseAlreadyResolved ResponseStatus = 4030
	StatusPromiseAlreadyRejected ResponseStatus = 4031
	StatusPromiseAlreadyCanceled ResponseStatus = 4032
	StatusPromiseAlreadyTimedOut ResponseStatus = 4033
	StatusPromiseNotFound        ResponseStatus = 4040
	StatusSubscriptionNotFound   ResponseStatus = 4041
	StatusPromiseAlreadyExists   ResponseStatus = 4090
)

// Platform level errors (5000-5999)

type ResonateErrorCode int

func (e ResonateErrorCode) String() string {
	return strconv.Itoa(int(e))
}

const (
	// catch call for now
	ErrInternalServer = iota + 5000

	// API
	ErrSystemShuttingDown
	ErrAPISubmissionQueueFull

	// AIO
	ErrAIOSubmissionQueueFull
	ErrAIONetworkFailure
	ErrAIOStoreFailure
	ErrAIOStoreSerializationFailure
)

type ResonateError struct {
	code   ResonateErrorCode
	reason string
	ogErr  error
}

func NewResonateError(code ResonateErrorCode, out string, in error) *ResonateError {
	return &ResonateError{
		code:   code,
		reason: out,
		ogErr:  in,
	}
}

func (e *ResonateError) Error() string {
	return e.reason
}

func (e *ResonateError) Unwrap() error {
	return e.ogErr
}

func (e *ResonateError) Code() ResonateErrorCode {
	return e.code
}

func (e *ResonateError) GRPC() code.Code {
	switch e.code {
	case ErrInternalServer:
		return code.Code_INTERNAL
	case ErrSystemShuttingDown:
		return code.Code_UNAVAILABLE
	case ErrAPISubmissionQueueFull:
		return code.Code_UNAVAILABLE
	case ErrAIOSubmissionQueueFull:
		return code.Code_UNAVAILABLE
	case ErrAIONetworkFailure:
		return code.Code_UNAVAILABLE
	case ErrAIOStoreFailure:
		return code.Code_UNAVAILABLE
	case ErrAIOStoreSerializationFailure:
		return code.Code_UNAVAILABLE
	default:
		return code.Code_UNKNOWN
	}
}
