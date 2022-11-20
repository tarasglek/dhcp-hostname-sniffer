This application sniffs the network traffic using libpcap for DHCP requests. It then checks if the ip sending dhcp requests has a `http://<ip>/metrics` endpoint.

This is useful to passively enumerate prometheus endpoints as they appear on the network.


for prod `make` creates a static executable.

for dev:

`go run . -f dhcp.dump`


poolpump cc:50:e3:68:9b:df