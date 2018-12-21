FROM ubuntu:18.04

MAINTAINER Gustavo Maurizio <gustavo.maurizio@gmail.com>  

ENV GOVERSION 1.11.4
ENV GOROOT /usr/local/go
ENV GOPATH /mnt
ENV PATH="/usr/local/go/bin:${PATH}"

RUN	apt-get update && \
	apt-get install -y gcc file git curl wget iperf3 bmon atop htop iftop && \
	curl -O https://storage.googleapis.com/golang/go${GOVERSION}.linux-amd64.tar.gz && \
	tar -xvf go1.11.4.linux-amd64.tar.gz && mv go /usr/local

CMD ["/bin/bash"]