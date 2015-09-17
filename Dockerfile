FROM golang:1.5.1-wheezy
RUN mkdir -p /opt/go
ENV GOPATH /opt/go
RUN go get github.com/mailgun/oxy/forward
RUN go get github.com/mailgun/oxy/testutils
RUN go get github.com/dockerx/dockerbox-proxy
RUN cd $GOPATH/src/github.com/dockerx/dockerbox-proxy && go build && cp ./dockerbox-proxy /usr/bin/
EXPOSE 9090 80
CMD ["/usr/bin/dockerbox-proxy"]
