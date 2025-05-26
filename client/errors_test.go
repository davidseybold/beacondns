package client

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/davidseybold/beacondns/internal/beaconerr"
)

func TestParseError(t *testing.T) {
	tests := []struct {
		name       string
		errResp    errorResponse
		wantErr    error
		wantErrMsg string
	}{
		{
			name: "zone already exists",
			errResp: errorResponse{
				Code:    string(beaconerr.ErrorCodeZoneAlreadyExists),
				Message: "zone already exists",
			},
			wantErr:    &ZoneAlreadyExistsError{},
			wantErrMsg: "ZoneAlreadyExists: zone already exists",
		},
		{
			name: "no such zone",
			errResp: errorResponse{
				Code:    string(beaconerr.ErrorCodeNoSuchZone),
				Message: "zone not found",
			},
			wantErr:    &NoSuchZoneError{},
			wantErrMsg: "NoSuchZone: zone not found",
		},
		{
			name: "response policy already exists",
			errResp: errorResponse{
				Code:    string(beaconerr.ErrorCodeResponsePolicyAlreadyExists),
				Message: "response policy already exists",
			},
			wantErr:    &ResponsePolicyAlreadyExistsError{},
			wantErrMsg: "ResponsePolicyAlreadyExists: response policy already exists",
		},
		{
			name: "no such response policy",
			errResp: errorResponse{
				Code:    string(beaconerr.ErrorCodeNoSuchResponsePolicy),
				Message: "response policy not found",
			},
			wantErr:    &NoSuchResponsePolicyError{},
			wantErrMsg: "NoSuchResponsePolicy: response policy not found",
		},
		{
			name: "no such response policy rule",
			errResp: errorResponse{
				Code:    string(beaconerr.ErrorCodeNoSuchResponsePolicyRule),
				Message: "response policy rule not found",
			},
			wantErr:    &NoSuchResponsePolicyRuleError{},
			wantErrMsg: "NoSuchResponsePolicyRule: response policy rule not found",
		},
		{
			name: "no such change",
			errResp: errorResponse{
				Code:    string(beaconerr.ErrorCodeNoSuchChange),
				Message: "change not found",
			},
			wantErr:    &NoSuchChangeError{},
			wantErrMsg: "NoSuchChange: change not found",
		},
		{
			name: "no such resource record set",
			errResp: errorResponse{
				Code:    string(beaconerr.ErrorCodeNoSuchResourceRecordSet),
				Message: "resource record set not found",
			},
			wantErr:    &NoSuchResourceRecordSetError{},
			wantErrMsg: "NoSuchResourceRecordSet: resource record set not found",
		},
		{
			name: "invalid argument",
			errResp: errorResponse{
				Code:    string(beaconerr.ErrorCodeInvalidArgument),
				Message: "invalid argument",
			},
			wantErr:    &InvalidArgumentError{},
			wantErrMsg: "InvalidArgument: invalid argument",
		},
		{
			name: "internal error",
			errResp: errorResponse{
				Code:    string(beaconerr.ErrorCodeInternalError),
				Message: "internal error",
			},
			wantErr:    &InternalError{},
			wantErrMsg: "InternalError: internal error",
		},
		{
			name: "unknown error code",
			errResp: errorResponse{
				Code:    "UnknownError",
				Message: "unknown error",
			},
			wantErr:    &beaconError{},
			wantErrMsg: "UnknownError: unknown error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := parseError(tt.errResp)
			assert.IsType(t, tt.wantErr, err)
			assert.Equal(t, tt.wantErrMsg, err.Error())
		})
	}
}
