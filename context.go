package main

const (
	AType     = "A"
	CNAMEType = "CNAME"
)

type Context struct {
	Config struct {
		Origin string
	}

	Records Records
}

type Record struct {
	Host string
	Type string
	TTL  uint32
	Data string
}

type Records []Record

func (rs Records) ARecords() Records {
	ret := Records{}
	for _, r := range rs {
		if r.Type == AType {
			ret = append(ret, r)
		}
	}
	return ret
}

func (rs Records) CNAMERecords() Records {
	ret := Records{}
	for _, r := range rs {
		if r.Type == CNAMEType {
			ret = append(ret, r)
		}
	}
	return ret
}
