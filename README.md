## Fuzzy DNS

A very simple domain name server that fuzzily matches its records. This fuzzy matching tolerates typos in the subdomains. For example, if the subdomain name is "google.example.com", opening (gogle.example.com, googl.example.com or ggl.example.com) will all work.

***Currently only A and CNAME records are supported.***

### Why?

Because it's fun. I googled and I didn't find something like this, so I implemented it. Is it useful? Honestly, I don't know.

### Configuration

The sample configuration contains :

```yaml
# The domain suffix to be trimmed
origin: "example.com"

# The bind address
address: "0.0.0.0:5333"

records:
      # The subdomain
    - host: "google"
      # Time to live
      ttl: 0
      # Record type
      type: "A"
      # Record Value
      data: "1.1.1.1"
    - host: "willfind"
      ttl: 0
      type: "A"
      data: "1.1.1.2"
    - host: "wontfind"
      ttl: 0
      type: "CNAME"
      # Notice the trailing dot
      data: "willfind.example.com."
```


### Deployment

#### Locally For testing

If you want to deploy it for fun. You will have to :

```bash
# Install it
go get github.com/MohamedBassem/fuzzy-dns

# Copy the sample config and modify it if needed
cp config.yml.sample config.yml

# Run it (Assuming that $GOPATH/bin is in your path)
fuzzy-dns --config config.yml --verbose
```

We will test it in another terminal. Those examples work on the sample config mentioned above.

```bash
# A normal query asking for the A record
$ dig +noall +answer @localhost -p 5333 A google.example.com
google.example.com.     0       IN      A       1.1.1.1

# Query with a typo.
$ dig +noall +answer @localhost -p 5333 A gogle.example.com
gogle.example.com.      0       IN      A       1.1.1.1

# Query with another typo
$ dig +noall +answer @localhost -p 5333 A ggl.example.com
ggl.example.com.        0       IN      A       1.1.1.1

# If there aren't any matches in the A records, CNAME records are matched
# According to RFC 1035.
$ dig +noall +answer @localhost -p 5333 A wontfind.example.com
wontfind.example.com.   0       IN      CNAME   willfind.example.com.
willfind.example.com.   0       IN      A       1.1.1.2

# Same as the previous example but with a typo
$ dig +noall +answer @localhost -p 5333 A wontfd.example.com
wontfd.example.com.     0       IN      CNAME   willfind.example.com.
willfind.example.com.   0       IN      A       1.1.1.2

# Asking for the CNAME record instead of the A record
$ dig +noall +answer @localhost -p 5333 CNAME wontfind.example.com
wontfind.example.com.   0       IN      CNAME   willfind.example.com.
```

### Operation

If the query is asking for an A record. A records are fuzzily searched. If the server couldn't find any matches in the A records, CNAME records are then fuzzily searched. This behaviour is similar to what's mentioned in RFC 1035.

Quoting RFC 1035 :
> "CNAME RRs cause special action in DNS software.  When a name server
> fails to find a desired RR in the resource set associated with the
> domain name, it checks to see if the resource set consists of a CNAME
> record with a matching class.  If so, the name server includes the CNAME
> record in the response and restarts the query at the domain name
> specified in the data field of the CNAME record.  The one exception to
> this rule is that queries which match the CNAME type are not restarted."


### TODO

- [ ] MX. It will be fun to tolerate typos in email domains
- [ ] AAAA records
- [ ] Tests

##Contribution
Your contributions and ideas are welcomed through issues and pull requests.

##License
Copyright (c) 2015, Mohamed Bassem. (MIT License)

See LICENSE for more info.
