package main

import (
	"fmt"
	"io/ioutil"
	"net"

	"github.com/miekg/dns"
	"gopkg.in/yaml.v2"
)

const (
	// AType is the A record string in the config
	AType = "A"
	// CNAMEType is the CNAME record string in the config
	CNAMEType = "CNAME"
)

// A Context struct is one of the attributes of the Server struct. It carrys the global configuration for the server as well as the DNS records
type Context struct {
	Origin  string
	Address string

	Records Records
}

// NewContextFromFile function reads and parses the config file.
func NewContextFromFile(filename string) (*Context, error) {

	file, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	ctx := Context{}
	err = yaml.Unmarshal(file, &ctx)
	if err != nil {
		return nil, err
	}

	if err := ctx.Records.validateAndNormalizeRecords(); err != nil {
		return nil, err
	}

	return &ctx, nil
}

// Record struct represents a single DNS record.
type Record struct {
	// The subdomain
	Host string

	// Record type. Check the package constants
	Type string

	// Time to live
	TTL uint32

	// The record value
	Data string
}

// Records type represents a set of records
type Records []Record

// ARecords returns all the records of type A.
func (rs Records) ARecords() Records {
	ret := Records{}
	for _, r := range rs {
		if r.Type == AType {
			ret = append(ret, r)
		}
	}
	return ret
}

// CNAMERecords returns all the records of type CNAME.
func (rs Records) CNAMERecords() Records {
	ret := Records{}
	for _, r := range rs {
		if r.Type == CNAMEType {
			ret = append(ret, r)
		}
	}
	return ret
}

func (rs Records) validateAndNormalizeRecords() error {
	for i := range rs {

		if rs[i].Host == "" {
			return fmt.Errorf("Host cannot be empty")
		}

		if rs[i].Host == "@" {
			return fmt.Errorf("@ is currently not a supported host")
		}

		switch rs[i].Type {

		case AType:
			ip := net.ParseIP(rs[i].Data)
			if ip == nil {
				return fmt.Errorf("Invalid IP '%v' in host %v", rs[i].Data, rs[i].Host)
			}
			if len(ip) != 4 && len(ip) != 16 {
				return fmt.Errorf("A records must have an IPv4 : '%v' in host %v", rs[i].Data, rs[i].Host)
			}

		case CNAMEType:
			if _, ok := dns.IsDomainName(rs[i].Data); !ok {
				return fmt.Errorf("%v is not a valid domain name", rs[i].Data)
			}
			rs[i].Data = dns.Fqdn(rs[i].Data)

		default:
			return fmt.Errorf("Invalid/Unsupported record type %v", rs[i].Type)

		}
	}
	return nil
}
