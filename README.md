## Fuzzy DNS

A very simple domain name server that fuzzily matches its records. This fuzzy matching tolerates typos in the subdomains. For example, if the subdomain name is "google.example.com", opening (gogle.example.com, googl.example.com or ggl.example.com) will all work.

***Currently only A and CNAME records are supported.***

*It's not a public DNS server, it's a privately managed DNS server. You get control on the domains to resolve. Check the actual deployment section.*

### Why?

Because it's fun. I googled and I didn't find something like this, so I implemented it. I learned [some things](http://blog.mbassem.com/2016/06/14/fuzzy-dns/) also while building it. Is it useful? Honestly, I don't know.

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
      data: "willfind.example.com"
```


### Deployment

#### Locally For testing

If you want to deploy it for fun. You will have to :

```bash
# Install it
go get github.com/MohamedBassem/fuzzy-dns

# Copy the sample config and modify it if needed
cp $GOPATH/src/github.com/MohamedBassem/fuzzy-dns/config.yml.sample config.yml

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

# Querying another domain
$ dig +noall +answer @localhost -p 5333 A willfind.example.com
willfind.example.com.   0       IN      A       1.1.1.2

# With a typo
$ dig +noall +answer @localhost -p 5333 A wllfid.example.com
willfid.example.com.    0       IN      A       1.1.1.2

# If there aren't any matches in the A records, CNAME records are matched
# According to RFC 1034.
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

And here's the demo

[![asciicast](https://asciinema.org/a/48805.png)](https://asciinema.org/a/48805)

#### Actual Deployment

Let's say we own the domain `example.com`. If you want to resolve the subdomains of `fuzzy.example.com` fuzzily you should do the following:

*Note, it can also be done with the subdomains of "example.com" directly, but you probably don't want to do this while the server is still in beta*


- Start a publicly accessible machine (on DigitalOcean, AWS, ..).
- Add an `NS` record in your domain's zone file (e.g. on Godaddy or whatever) with the host `fuzzy` and value `fuzzyns.example.com`.
- Add another `A` record in your zone file with the host `fuzzyns` and the value is the IP of the public server you created.
- By now, the resolution of the subdomain `fuzzy` is delegated to the server.
- Now on the server you created, pull fuzzy-dns using `go get github.com/MohamedBassem/fuzzy-dns`.
- Copy the sample config `cp $GOPATH/src/github.com/MohamedBassem/fuzzy-dns/config.yml.sample config.yml`
- Do the following changes to the config file:
  - Change the origin to `fuzzy.example.com`.
  - Change the address to `0.0.0.0:53`.
  - Change the records to some that actually make sense.
- Start the server with `$GOPATH/bin/fuzzy-dns --config config.yml --verbose`. You must start the server as root to be able to bind to port 53.
- In your browser try accessing `google.fuzzy.example.com` or `gogle.fuzzy.example.com` and they will all open the same IP you configured.

*You should replace "example.com" with your actual domain in all the previous examples.*


### Operation

If the query is asking for an A record. A records are fuzzily searched. If the server couldn't find any matches in the A records, CNAME records are then fuzzily searched. This behaviour is similar to what's mentioned in RFC 1034.

Quoting RFC 1034 :
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
