dhcp-hostname-sniffer: Dockerfile *.go
	docker build -t  dhcp-hostname-sniffer .
	docker run --rm dhcp-hostname-sniffer tar  -C /app -c dhcp-hostname-sniffer | tar xv
clean:
	rm -f dhcp-hostname-sniffer