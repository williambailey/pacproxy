FROM golang:1.10.3

WORKDIR /go/src/github.com/shakirshakiel/pacproxy
COPY . .
ENV CONFIG $CONFIG
RUN make
CMD /go/src/github.com/shakirshakiel/pacproxy/bin/pacproxy -c $CONFIG -v -l $HOSTNAME:8080
