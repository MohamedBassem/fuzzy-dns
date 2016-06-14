package main

import (
	"io/ioutil"

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
