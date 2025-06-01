package client

import (
	"fmt"

	"github.com/davidseybold/beacondns/internal/beaconerr"
)

type beaconError struct {
	Code    string
	Message string
}

func (e *beaconError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *beaconError) ErrorCode() string {
	return e.Code
}

func (e *beaconError) ErrorMessage() string {
	return e.Message
}

type ZoneAlreadyExistsError struct {
	beaconError
}

type NoSuchZoneError struct {
	beaconError
}

type ResponsePolicyAlreadyExistsError struct {
	beaconError
}

type ResponsePolicyRuleAlreadyExistsError struct {
	beaconError
}

type NoSuchResponsePolicyError struct {
	beaconError
}

type NoSuchResponsePolicyRuleError struct {
	beaconError
}

type NoSuchChangeError struct {
	beaconError
}

type NoSuchResourceRecordSetError struct {
	beaconError
}

type HostedZoneNotEmptyError struct {
	beaconError
}

type InvalidArgumentError struct {
	beaconError
}

type InternalError struct {
	beaconError
}

type NoSuchDomainListError struct {
	beaconError
}

type NoSuchFirewallRuleError struct {
	beaconError
}

type DomainExistsInDomainListError struct {
	beaconError
}

func parseError(errResponse errorResponse) error {
	bErr := beaconError(errResponse)
	code := beaconerr.ErrorCode(errResponse.Code)
	switch code {
	case beaconerr.ErrorCodeZoneAlreadyExists:
		return &ZoneAlreadyExistsError{
			beaconError: bErr,
		}
	case beaconerr.ErrorCodeNoSuchZone:
		return &NoSuchZoneError{
			beaconError: bErr,
		}
	case beaconerr.ErrorCodeNoSuchChange:
		return &NoSuchChangeError{
			beaconError: bErr,
		}
	case beaconerr.ErrorCodeNoSuchResourceRecordSet:
		return &NoSuchResourceRecordSetError{
			beaconError: bErr,
		}
	case beaconerr.ErrorCodeInvalidArgument:
		return &InvalidArgumentError{
			beaconError: bErr,
		}
	case beaconerr.ErrorCodeInternalError:
		return &InternalError{
			beaconError: bErr,
		}
	case beaconerr.ErrorCodeHostedZoneNotEmpty:
		return &HostedZoneNotEmptyError{
			beaconError: bErr,
		}
	case beaconerr.ErrorCodeNoSuchDomainList:
		return &NoSuchDomainListError{
			beaconError: bErr,
		}
	case beaconerr.ErrorCodeNoSuchFirewallRule:
		return &NoSuchFirewallRuleError{
			beaconError: bErr,
		}
	case beaconerr.ErrorCodeDomainExistsInDomainList:
		return &DomainExistsInDomainListError{
			beaconError: bErr,
		}
	default:
		return &bErr
	}
}
