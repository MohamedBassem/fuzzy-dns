package main

import (
	"fmt"
	"net"
	"sort"
	"strings"

	"github.com/miekg/dns"
	"github.com/renstrom/fuzzysearch/fuzzy"
)

type Server struct {
	ctx *Context
}

func (s *Server) trimOrigin(name string) string {
	return strings.TrimSuffix(name, s.ctx.Config.Origin)
}

func (s *Server) fuzzyMatchHost(name string, rs Records) Records {

	hosts := make([]string, len(rs))
	for i := range rs {
		hosts[i] = rs[i].Host
	}

	ranks := fuzzy.RankFindFold(name, hosts)
	sort.Sort(ranks)

	if len(ranks) == 0 {
		return Records{}
	}

	ret := Records{}
	for _, r := range rs {
		if r.Host == ranks[0].Source {
			ret = append(ret, r)
		}
	}
	return ret
}

func (s *Server) handleARecords(name string) []dns.RR {
	as := s.ctx.Records.ARecords()
	matches := s.fuzzyMatchHost(name, as)
	ret := []dns.RR{}
	for _, m := range matches {
		ret = append(ret, &dns.A{
			Hdr: dns.RR_Header{
				Name:   name,
				Rrtype: dns.TypeA,
				Class:  dns.ClassINET,
				Ttl:    m.TTL,
			},
			A: net.ParseIP(m.Data),
		})
	}
	return ret
}

func (s *Server) handleCNAMERecords(name string) []dns.RR {
	cs := s.ctx.Records.CNAMERecords()
	matches := s.fuzzyMatchHost(name, cs)
	ret := []dns.RR{}
	for _, m := range matches {
		ret = append(ret, &dns.CNAME{
			Hdr: dns.RR_Header{
				Name:   name,
				Rrtype: dns.TypeCNAME,
				Class:  dns.ClassINET,
				Ttl:    m.TTL,
			},
			Target: m.Data,
		})
	}

	// According to RFC1035, an exact match on the whole zonefile for the
	// A record should be done.
	as := s.ctx.Records.ARecords()
	for _, c := range matches {
		name := s.trimOrigin(c.Data)
		for _, a := range as {
			if a.Host == name {
				ret = append(ret, &dns.A{
					Hdr: dns.RR_Header{
						Name:   name,
						Rrtype: dns.TypeA,
						Class:  dns.ClassINET,
						Ttl:    a.TTL,
					},
					A: net.ParseIP(a.Data),
				})
			}
		}
	}

	return ret
}

func (s *Server) handleQuestion(q dns.Question) []dns.RR {

	switch q.Qtype {

	case dns.TypeA:
		as := s.handleARecords(s.trimOrigin(q.Name))
		if as == nil || len(as) == 0 {
			return s.handleCNAMERecords(q.Name)
		} else {
			return as
		}

	case dns.TypeCNAME:
		return s.handleCNAMERecords(q.Name)

	default:
		return nil
	}

}

func (s *Server) HandleRequest(w dns.ResponseWriter, r *dns.Msg) {
	resp := &dns.Msg{}
	resp.SetReply(r)

	for _, q := range r.Question {
		ans := s.handleQuestion(q)
		if ans != nil {
			resp.Answer = append(resp.Answer, ans...)
		}
	}

	w.WriteMsg(resp)
	w.Close()

}

func main() {
	s := Server{}
	dns.HandleFunc(".", s.HandleRequest)
	server := &dns.Server{Addr: "0.0.0.0:5333", Net: "udp"}
	err := server.ListenAndServe()
	fmt.Println(err)
}
