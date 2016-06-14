# dns-fuzz


### Operation

Quoting RFC 1035 :
"CNAME RRs cause special action in DNS software.  When a name server
fails to find a desired RR in the resource set associated with the
domain name, it checks to see if the resource set consists of a CNAME
record with a matching class.  If so, the name server includes the CNAME
record in the response and restarts the query at the domain name
specified in the data field of the CNAME record.  The one exception to
this rule is that queries which match the CNAME type are not restarted."


### TODO
- MX : it wouldn't it be fun to tolerate typos in email domains?
