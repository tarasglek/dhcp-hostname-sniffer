# docker build -t  dhcp-hostname-sniffer .
FROM golang:1.19

MAINTAINER Marc Lallaouret <mlallaouret@gmail.com> 

RUN apt-get update \
    && apt-get install flex bison curl -y \
    && apt-get clean

ENV PCAP_VERSION 1.10.1
RUN wget http://www.tcpdump.org/release/libpcap-$PCAP_VERSION.tar.gz && tar xzf libpcap-*.tar.gz \
    && cd libpcap-* \
    && ./configure && make install

WORKDIR /app

COPY go.mod ./
COPY go.sum ./

RUN go mod download

COPY *.go ./

# RUN go build

RUN go build --ldflags " -linkmode external -extldflags \"-static\"" 