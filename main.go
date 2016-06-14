package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"sort"
	"strings"

	"github.com/miekg/dns"
	"github.com/renstrom/fuzzysearch/fuzzy"
)

// Server struct is the main server itself carrying the global context and the logger
type Server struct {
	ctx    *Context
	logger *log.Logger
}

// NewServer creates a new server instance
func NewServer(ctx *Context, logger *log.Logger) *Server {
	return &Server{
		ctx:    ctx,
		logger: logger,
	}
}

func (s *Server) trimOrigin(name string) string {
	return strings.TrimSuffix(name, "."+s.ctx.Origin+".")
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
		if r.Host == ranks[0].Target {
			ret = append(ret, r)
		}
	}
	return ret
}

func (s *Server) handleARecords(originalName string) []dns.RR {
	name := s.trimOrigin(originalName)
	as := s.ctx.Records.ARecords()
	matches := s.fuzzyMatchHost(name, as)
	ret := []dns.RR{}
	for _, m := range matches {
		ret = append(ret, &dns.A{
			Hdr: dns.RR_Header{
				Name:   originalName,
				Rrtype: dns.TypeA,
				Class:  dns.ClassINET,
				Ttl:    m.TTL,
			},
			A: net.ParseIP(m.Data),
		})
	}
	return ret
}

func (s *Server) handleCNAMERecords(originalName string, rec bool) []dns.RR {
	name := s.trimOrigin(originalName)
	cs := s.ctx.Records.CNAMERecords()
	matches := s.fuzzyMatchHost(name, cs)
	ret := []dns.RR{}
	for _, m := range matches {
		ret = append(ret, &dns.CNAME{
			Hdr: dns.RR_Header{
				Name:   originalName,
				Rrtype: dns.TypeCNAME,
				Class:  dns.ClassINET,
				Ttl:    m.TTL,
			},
			Target: m.Data,
		})
	}

	// According to RFC1035, an exact match on the whole zonefile for the
	// A record should be done.
	if rec {
		as := s.ctx.Records.ARecords()
		for _, c := range matches {
			oname := c.Data
			name := s.trimOrigin(oname)
			for _, a := range as {
				if a.Host == name {
					ret = append(ret, &dns.A{
						Hdr: dns.RR_Header{
							Name:   oname,
							Rrtype: dns.TypeA,
							Class:  dns.ClassINET,
							Ttl:    a.TTL,
						},
						A: net.ParseIP(a.Data),
					})
				}
			}
		}
	}

	return ret
}

func (s *Server) handleQuestion(q dns.Question) []dns.RR {

	switch q.Qtype {

	case dns.TypeA:
		as := s.handleARecords(q.Name)
		if as == nil || len(as) == 0 {
			return s.handleCNAMERecords(q.Name, true)
		}
		return as

	case dns.TypeCNAME:
		return s.handleCNAMERecords(q.Name, false)

	default:
		return nil
	}

}

// LoggedRequest is a DNS handler function to wrap the original handler with a query logger
func (s *Server) LoggedRequest(f dns.HandlerFunc) dns.HandlerFunc {

	return func(w dns.ResponseWriter, r *dns.Msg) {
		s.logger.Println(r.String())
		f(w, r)
	}
}

// HandleRequest is the main request handler. It recieves, parses and responds to the DNS queries.
func (s *Server) HandleRequest(w dns.ResponseWriter, r *dns.Msg) {
	resp := &dns.Msg{}
	resp.SetReply(r)

	for _, q := range r.Question {
		ans := s.handleQuestion(q)
		if ans != nil {
			resp.Answer = append(resp.Answer, ans...)
		}
	}

	err := w.WriteMsg(resp)
	if err != nil {
		s.logger.Println("ERROR : " + err.Error())
	}
	w.Close()

}

func main() {

	configFile := flag.String("config", "", "Config file path")
	verbose := flag.Bool("verbose", false, "Verbose log with each request")
	flag.Parse()

	logger := log.New(os.Stdout, "", log.LstdFlags)

	if *configFile == "" {
		fmt.Println("--config flag cannot be empty")
		os.Exit(1)
	}

	ctx, err := NewContextFromFile(*configFile)
	if err != nil {
		fmt.Println("Error: Failed parsing config: " + err.Error())
		os.Exit(1)
	}

	s := NewServer(ctx, logger)

	if *verbose {
		dns.HandleFunc(".", s.LoggedRequest(s.HandleRequest))
	} else {
		dns.HandleFunc(".", s.HandleRequest)
	}
	logger.Printf("Server listening to address: %v\n", s.ctx.Address)
	server := &dns.Server{Addr: s.ctx.Address, Net: "udp"}
	err = server.ListenAndServe()
	fmt.Println(err)
}
