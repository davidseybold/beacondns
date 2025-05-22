package parse

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"unicode"

	"github.com/davidseybold/beacondns/internal/model"
	"github.com/miekg/dns"
)

const (
	soaRecordFieldNumber           = 7
	caaRecordFieldNumber           = 3
	dsRecordFieldNumber            = 4
	httpsRecordFieldNumber         = 2
	naptrRecordFieldNumber         = 6
	srvRecordFieldNumber           = 4
	sshfpRecordFieldNumber         = 3
	tlsaRecordFieldNumber          = 4
	tlsaUsageMaxFieldNumber        = 3
	tlsaSelectorMaxFieldNumber     = 1
	tlsaMatchingTypeMaxFieldNumber = 2
)

var (
	ErrNoResourceRecords            = errors.New("no resource records")
	ErrInvalidInteger               = errors.New("value is not an integer")
	ErrInvalidIPv4Address           = errors.New("value is not a valid IPv4 address")
	ErrInvalidIPv6Address           = errors.New("value is not a valid IPv6 address")
	ErrInvalidDomainName            = errors.New("value is not a valid domain name")
	ErrSOARecordFieldCount          = errors.New("SOA record doesn't have 7 fields")
	ErrCAARecordFieldCount          = errors.New("CAA record doesn't have 3 fields")
	ErrDSRecordFieldCount           = errors.New("DS record doesn't have 4 fields")
	ErrSVCBRecordRequiredFieldCount = errors.New("SVCB record doesn't have 2 fields")
	ErrSVCBInvalidKeyValue          = errors.New("SVCB record has invalid key-value pair")
	ErrNAPTRRecordFieldCount        = errors.New("NAPTR record doesn't have 6 fields")
	ErrInvalidNAPTRRecordFlags      = errors.New("NAPTR record is invalid")
	ErrSRVRecordFieldCount          = errors.New("SRV record doesn't have 4 fields")
	ErrSSHFPRecordFieldCount        = errors.New("SSHFP record doesn't have 3 fields")
	ErrTLSARecordFieldCount         = errors.New("TLSA record doesn't have 4 fields")
	ErrInvalidMXFieldCount          = errors.New("MX record must have 2 fields")
	ErrInvalidTLSAField             = errors.New("TLSA record field is invalid")
	ErrInvalidTXTField              = errors.New("TXT record field is invalid")
	ErrInvalidSOASerial             = errors.New("SOA serial number must be in YYYYMMDDnn format")
	ErrInvalidMXPriority            = errors.New("MX priority must be between 0 and 65535")
	ErrInvalidSRVPort               = errors.New("SRV port must be between 0 and 65535")
	ErrInvalidSRVWeight             = errors.New("SRV weight must be between 0 and 65535")
	ErrInvalidSRVPriority           = errors.New("SRV priority must be between 0 and 65535")
	ErrInvalidNAPTRFlags            = errors.New("NAPTR flags must be quoted")
	ErrInvalidNAPTRService          = errors.New("NAPTR service must be quoted")
	ErrInvalidNAPTRRegexp           = errors.New("NAPTR regexp must be quoted")
	ErrInvalidTXTValue              = errors.New("TXT value must be quoted")
	ErrInvalidCAAValue              = errors.New("CAA value must be quoted")
	ErrInvalidDSDigest              = errors.New("DS digest must be a valid hex string")
	ErrInvalidTLSAUsage             = errors.New("TLSA usage must be between 0 and 3")
	ErrInvalidTLSASelector          = errors.New("TLSA selector must be between 0 and 1")
	ErrInvalidTLSAMatching          = errors.New("TLSA matching type must be between 0 and 2")
	ErrInvalidSSHFPAlgorithm        = errors.New("SSHFP algorithm must be between 0 and 4")
	ErrInvalidSSHFPType             = errors.New("SSHFP type must be between 0 and 2")
	ErrInvalidCharacterString       = errors.New("value should be enclosed in quotation marks")
	ErrInvalidMandatoryKey          = errors.New("invalid mandatory key")
)

func A(rrset model.ResourceRecordSet) ([]dns.RR, error) {
	if len(rrset.ResourceRecords) == 0 {
		return nil, ErrNoResourceRecords
	}

	dnsRRs := make([]dns.RR, 0, len(rrset.ResourceRecords))
	for _, rr := range rrset.ResourceRecords {
		r := new(dns.A)
		r.Hdr = createHeader(rrset.Name, dns.TypeA, rrset.TTL)
		ip := net.ParseIP(rr.Value)
		if ip == nil || ip.To4() == nil {
			return nil, valueError(ErrInvalidIPv4Address, rr.Value)
		}

		r.A = ip
		dnsRRs = append(dnsRRs, r)
	}

	return dnsRRs, nil
}

func AAAA(rrset model.ResourceRecordSet) ([]dns.RR, error) {
	if len(rrset.ResourceRecords) == 0 {
		return nil, ErrNoResourceRecords
	}

	dnsRRs := make([]dns.RR, 0, len(rrset.ResourceRecords))
	for _, rr := range rrset.ResourceRecords {
		r := new(dns.AAAA)
		r.Hdr = createHeader(rrset.Name, dns.TypeAAAA, rrset.TTL)
		ip := net.ParseIP(rr.Value)
		if ip == nil || ip.To16() == nil {
			return nil, valueError(ErrInvalidIPv6Address, rr.Value)
		}
		r.AAAA = ip
		dnsRRs = append(dnsRRs, r)
	}

	return dnsRRs, nil
}

func CAA(rrset model.ResourceRecordSet) ([]dns.RR, error) {
	if len(rrset.ResourceRecords) == 0 {
		return nil, ErrNoResourceRecords
	}

	dnsRRs := make([]dns.RR, 0, len(rrset.ResourceRecords))
	for _, rr := range rrset.ResourceRecords {
		r := new(dns.CAA)
		r.Hdr = createHeader(rrset.Name, dns.TypeCAA, rrset.TTL)

		parts := strings.Fields(rr.Value)
		if len(parts) != caaRecordFieldNumber {
			return nil, valueError(ErrCAARecordFieldCount, rr.Value)
		}

		flag, err := parse8BitUint(parts[0])
		if err != nil {
			return nil, err
		}

		value := parts[2]
		if !assertQuoted(value) {
			return nil, valueError(ErrInvalidCAAValue, value)
		}
		value = strings.Trim(value, "\"")

		r.Flag = flag
		r.Tag = parts[1]
		r.Value = value
		dnsRRs = append(dnsRRs, r)
	}

	return dnsRRs, nil
}

func CNAME(rrset model.ResourceRecordSet) ([]dns.RR, error) {
	if len(rrset.ResourceRecords) == 0 {
		return nil, ErrNoResourceRecords
	}

	dnsRRs := make([]dns.RR, 0, len(rrset.ResourceRecords))
	for _, rr := range rrset.ResourceRecords {

		r := new(dns.CNAME)
		r.Hdr = createHeader(rrset.Name, dns.TypeCNAME, rrset.TTL)

		r.Target = dns.Fqdn(rr.Value)
		dnsRRs = append(dnsRRs, r)
	}

	return dnsRRs, nil
}

func DS(rrset model.ResourceRecordSet) ([]dns.RR, error) {
	if len(rrset.ResourceRecords) == 0 {
		return nil, ErrNoResourceRecords
	}

	dnsRRs := make([]dns.RR, 0, len(rrset.ResourceRecords))
	for _, rr := range rrset.ResourceRecords {
		r := new(dns.DS)
		r.Hdr = createHeader(rrset.Name, dns.TypeDS, rrset.TTL)

		parts := strings.Fields(rr.Value)
		if len(parts) != dsRecordFieldNumber {
			return nil, valueError(ErrDSRecordFieldCount, rr.Value)
		}

		keyTag, err := parse16BitUint(parts[0])
		if err != nil {
			return nil, err
		}

		algorithm, err := parse8BitUint(parts[1])
		if err != nil {
			return nil, err
		}

		digestType, err := parse8BitUint(parts[2])
		if err != nil {
			return nil, err
		}

		r.KeyTag = keyTag
		r.Algorithm = algorithm
		r.DigestType = digestType
		r.Digest = parts[3]

		dnsRRs = append(dnsRRs, r)
	}

	return dnsRRs, nil
}

func HTTPS(rrset model.ResourceRecordSet) ([]dns.RR, error) {
	if len(rrset.ResourceRecords) == 0 {
		return nil, ErrNoResourceRecords
	}

	dnsRRs := make([]dns.RR, 0, len(rrset.ResourceRecords))
	for _, rr := range rrset.ResourceRecords {
		r := new(dns.HTTPS)
		r.Hdr = createHeader(rrset.Name, dns.TypeHTTPS, rrset.TTL)

		priority, target, keyValues, err := svcbValue(rr.Value)
		if err != nil {
			return nil, err
		}

		r.Priority = priority
		r.Target = target
		r.Value = keyValues

		dnsRRs = append(dnsRRs, r)
	}

	return dnsRRs, nil
}

func MX(rrset model.ResourceRecordSet) ([]dns.RR, error) {
	if len(rrset.ResourceRecords) == 0 {
		return nil, ErrNoResourceRecords
	}

	dnsRRs := make([]dns.RR, 0, len(rrset.ResourceRecords))
	for _, rr := range rrset.ResourceRecords {
		r := new(dns.MX)
		r.Hdr = createHeader(rrset.Name, dns.TypeMX, rrset.TTL)

		parts := strings.Fields(rr.Value)
		if len(parts) != 2 {
			return nil, valueError(ErrInvalidMXFieldCount, rr.Value)
		}

		priority, err := parse16BitUint(parts[0])
		if err != nil {
			return nil, err
		}

		r.Mx = dns.Fqdn(parts[1])
		r.Preference = priority
		dnsRRs = append(dnsRRs, r)
	}

	return dnsRRs, nil
}

func NAPTR(rrset model.ResourceRecordSet) ([]dns.RR, error) {
	if len(rrset.ResourceRecords) == 0 {
		return nil, ErrNoResourceRecords
	}

	dnsRRs := make([]dns.RR, 0, len(rrset.ResourceRecords))
	for _, rr := range rrset.ResourceRecords {
		r := new(dns.NAPTR)
		r.Hdr = createHeader(rrset.Name, dns.TypeNAPTR, rrset.TTL)

		parts := strings.Fields(rr.Value)
		if len(parts) != naptrRecordFieldNumber {
			return nil, valueError(ErrNAPTRRecordFieldCount, rr.Value)
		}

		order, err := parse16BitUint(parts[0])
		if err != nil {
			return nil, err
		}

		pref, err := parse16BitUint(parts[1])
		if err != nil {
			return nil, err
		}

		flags := parts[2]
		if !assertQuoted(flags) {
			return nil, valueError(ErrInvalidNAPTRRecordFlags, rr.Value)
		}
		flags = strings.Trim(flags, "\"")

		service := parts[3]
		if !assertQuoted(service) {
			return nil, valueError(ErrInvalidNAPTRRecordFlags, rr.Value)
		}
		service = strings.Trim(service, "\"")

		regexp := parts[4]
		if !assertQuoted(regexp) {
			return nil, valueError(ErrInvalidNAPTRRecordFlags, rr.Value)
		}
		regexp = strings.Trim(regexp, "\"")

		replacement := parts[5]

		r.Order = uint16(order)
		r.Preference = uint16(pref)
		r.Flags = flags
		r.Service = service
		r.Regexp = regexp
		r.Replacement = replacement

		dnsRRs = append(dnsRRs, r)
	}

	return dnsRRs, nil
}

func assertQuoted(value string) bool {
	return strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"")
}

func NS(rrset model.ResourceRecordSet) ([]dns.RR, error) {
	if len(rrset.ResourceRecords) == 0 {
		return nil, ErrNoResourceRecords
	}

	dnsRRs := make([]dns.RR, 0, len(rrset.ResourceRecords))
	for _, rr := range rrset.ResourceRecords {
		r := new(dns.NS)
		r.Hdr = createHeader(rrset.Name, dns.TypeNS, rrset.TTL)
		r.Ns = dns.Fqdn(rr.Value)
		dnsRRs = append(dnsRRs, r)
	}

	return dnsRRs, nil
}

func PTR(rrset model.ResourceRecordSet) ([]dns.RR, error) {
	if len(rrset.ResourceRecords) == 0 {
		return nil, ErrNoResourceRecords
	}

	dnsRRs := make([]dns.RR, 0, len(rrset.ResourceRecords))
	for _, rr := range rrset.ResourceRecords {
		r := new(dns.PTR)
		r.Hdr = createHeader(rrset.Name, dns.TypePTR, rrset.TTL)
		r.Ptr = dns.Fqdn(rr.Value)
		dnsRRs = append(dnsRRs, r)
	}

	return dnsRRs, nil
}

func SOA(rrset model.ResourceRecordSet) ([]dns.RR, error) {
	if len(rrset.ResourceRecords) == 0 {
		return nil, ErrNoResourceRecords
	}

	rr := rrset.ResourceRecords[0]
	r := new(dns.SOA)
	r.Hdr = createHeader(rrset.Name, dns.TypeSOA, rrset.TTL)

	parts := strings.Fields(rr.Value)
	if len(parts) != soaRecordFieldNumber {
		return nil, valueError(ErrSOARecordFieldCount, rr.Value)
	}

	r.Ns = dns.Fqdn(parts[0])
	r.Mbox = dns.Fqdn(parts[1])

	serial, err := parse32BitUint(parts[2])
	if err != nil {
		return nil, err
	}

	r.Serial = serial

	refresh, err := parse32BitUint(parts[3])
	if err != nil {
		return nil, err
	}
	r.Refresh = refresh

	retry, err := parse32BitUint(parts[4])
	if err != nil {
		return nil, err
	}
	r.Retry = retry

	expire, err := parse32BitUint(parts[5])
	if err != nil {
		return nil, err
	}
	r.Expire = expire

	minttl, err := parse32BitUint(parts[6])
	if err != nil {
		return nil, err
	}
	r.Minttl = minttl

	return []dns.RR{r}, nil
}

func SRV(rrset model.ResourceRecordSet) ([]dns.RR, error) {
	if len(rrset.ResourceRecords) == 0 {
		return nil, ErrNoResourceRecords
	}

	dnsRRs := make([]dns.RR, 0, len(rrset.ResourceRecords))
	for _, rr := range rrset.ResourceRecords {
		r := new(dns.SRV)
		r.Hdr = createHeader(rrset.Name, dns.TypeSRV, rrset.TTL)

		parts := strings.Fields(rr.Value)
		if len(parts) != srvRecordFieldNumber {
			return nil, valueError(ErrSRVRecordFieldCount, rr.Value)
		}

		priority, err := parse16BitUint(parts[0])
		if err != nil {
			return nil, err
		}

		weight, err := parse16BitUint(parts[1])
		if err != nil {
			return nil, err
		}

		port, err := parse16BitUint(parts[2])
		if err != nil {
			return nil, err
		}

		target := parts[3]
		if _, ok := dns.IsDomainName(target); !ok {
			return nil, valueError(ErrInvalidDomainName, target)
		}

		r.Priority = priority
		r.Weight = weight
		r.Port = port
		r.Target = target

		dnsRRs = append(dnsRRs, r)
	}

	return dnsRRs, nil
}

func SSHFP(rrset model.ResourceRecordSet) ([]dns.RR, error) {
	if len(rrset.ResourceRecords) == 0 {
		return nil, ErrNoResourceRecords
	}

	dnsRRs := make([]dns.RR, 0, len(rrset.ResourceRecords))
	for _, rr := range rrset.ResourceRecords {
		r := new(dns.SSHFP)
		r.Hdr = createHeader(rrset.Name, dns.TypeSSHFP, rrset.TTL)

		parts := strings.Fields(rr.Value)
		if len(parts) != sshfpRecordFieldNumber {
			return nil, valueError(ErrSSHFPRecordFieldCount, rr.Value)
		}

		algorithm, err := parse8BitUint(parts[0])
		if err != nil {
			return nil, err
		}

		hashType, err := parse8BitUint(parts[1])
		if err != nil {
			return nil, err
		}

		fingerprint := parts[2]

		r.Algorithm = algorithm
		r.Type = hashType
		r.FingerPrint = fingerprint

		dnsRRs = append(dnsRRs, r)
	}

	return dnsRRs, nil
}

func SVCB(rrset model.ResourceRecordSet) ([]dns.RR, error) {
	if len(rrset.ResourceRecords) == 0 {
		return nil, ErrNoResourceRecords
	}

	dnsRRs := make([]dns.RR, 0, len(rrset.ResourceRecords))
	for _, rr := range rrset.ResourceRecords {
		r := new(dns.SVCB)
		r.Hdr = createHeader(rrset.Name, dns.TypeSVCB, rrset.TTL)

		priority, target, keyValues, err := svcbValue(rr.Value)
		if err != nil {
			return nil, err
		}

		r.Priority = priority
		r.Target = target
		r.Value = keyValues

		dnsRRs = append(dnsRRs, r)
	}

	return dnsRRs, nil
}

var svcbStringToKeyMap = map[string]dns.SVCBKey{
	"mandatory":       dns.SVCB_MANDATORY,
	"alpn":            dns.SVCB_ALPN,
	"no-default-alpn": dns.SVCB_NO_DEFAULT_ALPN,
	"port":            dns.SVCB_PORT,
	"ipv4hint":        dns.SVCB_IPV4HINT,
	"ech":             dns.SVCB_ECHCONFIG,
	"ipv6hint":        dns.SVCB_IPV6HINT,
	"dohpath":         dns.SVCB_DOHPATH,
	"ohttp":           dns.SVCB_OHTTP,
}

func svcbValue(value string) (uint16, string, []dns.SVCBKeyValue, error) {
	parts := strings.Fields(value)
	if len(parts) < httpsRecordFieldNumber {
		return 0, "", nil, valueError(ErrSVCBRecordRequiredFieldCount, value)
	}

	priority, err := parse16BitUint(parts[0])
	if err != nil {
		return 0, "", nil, err
	}
	target := parts[1]
	keyValues := make([]dns.SVCBKeyValue, 0, len(parts)-2)

	for _, part := range parts[2:] {
		if part == "no-default-alpn" {
			keyValues = append(keyValues, &dns.SVCBNoDefaultAlpn{})
			continue
		}

		kvParts := strings.Split(part, "=")
		if len(kvParts) != 2 {
			return 0, "", nil, valueError(ErrSVCBInvalidKeyValue, value)
		}

		key := kvParts[0]
		value := kvParts[1]

		svcbKey, ok := svcbStringToKeyMap[key]
		if !ok {
			return 0, "", nil, valueError(ErrSVCBInvalidKeyValue, value)
		}

		switch svcbKey {
		case dns.SVCB_ALPN:
			s := strings.Trim(value, "\"")
			alpnParts := strings.Split(s, ",")
			alpn := make([]string, 0, len(alpnParts))
			for _, alpnPart := range alpnParts {
				alpn = append(alpn, strings.Trim(alpnPart, " "))
			}
			keyValues = append(keyValues, &dns.SVCBAlpn{
				Alpn: alpn,
			})
		case dns.SVCB_PORT:
			port, err := strconv.ParseUint(value, 10, 16)
			if err != nil {
				return 0, "", nil, valueError(ErrInvalidInteger, value)
			}
			keyValues = append(keyValues, &dns.SVCBPort{
				Port: uint16(port),
			})
		case dns.SVCB_IPV4HINT:
			tr := strings.Trim(value, "\"")
			strIps := strings.Split(tr, ",")
			ips := make([]net.IP, 0, len(strIps))
			for _, strIp := range strIps {
				ip := net.ParseIP(strIp)
				if ip == nil || ip.To4() == nil {
					return 0, "", nil, valueError(ErrInvalidIPv4Address, strIp)
				}
				ips = append(ips, ip)
			}
			keyValues = append(keyValues, &dns.SVCBIPv4Hint{
				Hint: ips,
			})
		case dns.SVCB_IPV6HINT:
			tr := strings.Trim(value, "\"")
			strIps := strings.Split(tr, ",")
			ips := make([]net.IP, 0, len(strIps))
			for _, strIp := range strIps {
				ip := net.ParseIP(strIp)
				if ip == nil || ip.To16() == nil {
					return 0, "", nil, valueError(ErrInvalidIPv6Address, strIp)
				}
				ips = append(ips, ip)
			}
			keyValues = append(keyValues, &dns.SVCBIPv6Hint{
				Hint: ips,
			})
		case dns.SVCB_DOHPATH:
			keyValues = append(keyValues, &dns.SVCBDoHPath{
				Template: value,
			})
		case dns.SVCB_ECHCONFIG:
			keyValues = append(keyValues, &dns.SVCBECHConfig{
				ECH: []byte(value),
			})
		case dns.SVCB_OHTTP:
			keyValues = append(keyValues, &dns.SVCBOhttp{})
		case dns.SVCB_MANDATORY:
			m := strings.Trim(value, "\"")
			mandatoryParts := strings.Split(m, ",")
			mandatoryCodes := make([]dns.SVCBKey, 0, len(mandatoryParts))
			for _, part := range mandatoryParts {
				code, ok := svcbStringToKeyMap[part]
				if !ok || code == dns.SVCB_MANDATORY {
					return 0, "", nil, valueError(ErrInvalidMandatoryKey, part)
				}
				mandatoryCodes = append(mandatoryCodes, code)
			}
			keyValues = append(keyValues, &dns.SVCBMandatory{Code: mandatoryCodes})
		}
	}

	return uint16(priority), target, keyValues, nil
}

func TLSA(rrset model.ResourceRecordSet) ([]dns.RR, error) {
	if len(rrset.ResourceRecords) == 0 {
		return nil, ErrNoResourceRecords
	}

	dnsRRs := make([]dns.RR, 0, len(rrset.ResourceRecords))
	for _, rr := range rrset.ResourceRecords {
		r := new(dns.TLSA)
		r.Hdr = createHeader(rrset.Name, dns.TypeTLSA, rrset.TTL)

		parts := strings.Fields(rr.Value)
		if len(parts) != tlsaRecordFieldNumber {
			return nil, valueError(ErrTLSARecordFieldCount, rr.Value)
		}

		usage, err := parse8BitUint(parts[0])
		if err != nil {
			return nil, err
		}
		if usage > tlsaUsageMaxFieldNumber {
			return nil, valueError(ErrInvalidTLSAField, parts[0])
		}

		selector, err := parse8BitUint(parts[1])
		if err != nil {
			return nil, err
		}
		if selector > tlsaSelectorMaxFieldNumber {
			return nil, valueError(ErrInvalidTLSAField, parts[1])
		}

		matchingType, err := parse8BitUint(parts[2])
		if err != nil {
			return nil, err
		}
		if matchingType > tlsaMatchingTypeMaxFieldNumber {
			return nil, valueError(ErrInvalidTLSAField, parts[2])
		}

		certificateAssociationData := parts[3]

		r.Usage = usage
		r.Selector = selector
		r.MatchingType = matchingType
		r.Certificate = certificateAssociationData

		dnsRRs = append(dnsRRs, r)
	}

	return dnsRRs, nil
}

func TXT(rrset model.ResourceRecordSet) ([]dns.RR, error) {
	if len(rrset.ResourceRecords) == 0 {
		return nil, ErrNoResourceRecords
	}

	var records []dns.RR
	for _, rr := range rrset.ResourceRecords {
		r := &dns.TXT{
			Hdr: dns.RR_Header{
				Name:   rrset.Name,
				Rrtype: dns.TypeTXT,
				Class:  dns.ClassINET,
				Ttl:    rrset.TTL,
			},
		}

		processedParts, err := txtValue(rr.Value)
		if err != nil {
			return nil, err
		}

		r.Txt = processedParts

		records = append(records, r)
	}

	return records, nil
}

func txtValue(value string) ([]string, error) {
	s := strings.TrimSpace(value)

	// Case: single unquoted string (no quotes at all)
	if !strings.Contains(s, `"`) {
		if strings.ContainsAny(s, "\n\r") || s == "" {
			return nil, valueError(ErrInvalidCharacterString, value)
		}
		return []string{s}, nil
	}

	// Quoted mode: must be entirely composed of quoted strings
	var results []string
	for len(s) > 0 {
		if s[0] != '"' {
			return nil, valueError(ErrInvalidCharacterString, value)
		}

		// Parse quoted string
		var buf strings.Builder
		escaped := false
		i := 1
		for i < len(s) {
			c := s[i]
			if escaped {
				buf.WriteByte(c)
				escaped = false
			} else if c == '\\' {
				escaped = true
			} else if c == '"' {
				break
			} else {
				buf.WriteByte(c)
			}
			i++
		}

		if i >= len(s) || s[i] != '"' {
			return nil, valueError(ErrInvalidCharacterString, value)
		}

		results = append(results, buf.String())
		s = strings.TrimLeftFunc(s[i+1:], unicode.IsSpace)
	}

	return results, nil

}

func valueError(err error, value string) error {
	return fmt.Errorf("(%w) encountered with '%s'", err, value)
}

func parse8BitUint(value string) (uint8, error) {
	parsed, err := strconv.ParseUint(value, 10, 8)
	if err != nil {
		return 0, valueError(ErrInvalidInteger, value)
	}
	return uint8(parsed), nil
}

func parse16BitUint(value string) (uint16, error) {
	parsed, err := strconv.ParseUint(value, 10, 16)
	if err != nil {
		return 0, valueError(ErrInvalidInteger, value)
	}
	return uint16(parsed), nil
}

func parse32BitUint(value string) (uint32, error) {
	parsed, err := strconv.ParseUint(value, 10, 32)
	if err != nil {
		return 0, valueError(ErrInvalidInteger, value)
	}
	return uint32(parsed), nil
}

func createHeader(name string, rrtype uint16, ttl uint32) dns.RR_Header {
	return dns.RR_Header{
		Name:   dns.Fqdn(name),
		Rrtype: rrtype,
		Class:  dns.ClassINET,
		Ttl:    ttl,
	}
}
