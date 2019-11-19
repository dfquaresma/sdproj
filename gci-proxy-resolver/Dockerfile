FROM openfaas/of-watchdog:0.7.2 as watchdog

FROM golang:1.11-alpine3.10 as builder

# Allows you to add additional packages via build-arg
ARG ADDITIONAL_PACKAGE

COPY --from=watchdog /fwatchdog /usr/bin/fwatchdog
RUN chmod +x /usr/bin/fwatchdog

RUN apk --no-cache add git

WORKDIR /go/src/github.com/dfquaresma
RUN git clone https://github.com/dfquaresma/sdproj.git \
     && cd sdproj/gci-proxy-resolver \
     && go get -v -u github.com/gorilla/mux \
     && go get -v -u github.com/docker/docker/api/types/swarm \
     && go build

FROM alpine:3.10
RUN apk --no-cache add \
    ca-certificates

# Add non root user
RUN addgroup -S app && adduser -S -g app app
RUN mkdir -p /home/app

WORKDIR /home/app

COPY --from=builder /usr/bin/fwatchdog                                                  .
COPY --from=builder /go/src/github.com/dfquaresma/sdproj/gci-proxy-resolver/gci-proxy-resolver     .

RUN chown -R app /home/app

USER app

ENV upstream_url="http://127.0.0.1:8082"
ENV mode="http"

ENV fprocess="./gci-proxy-resolver"
ENV MANAGER_ADDRESSES="150.165.75.121:5555,150.165.75.125:5555"
EXPOSE 8080

HEALTHCHECK --interval=3s CMD [ -e /tmp/.lock ] || exit 1

CMD ["./fwatchdog"]
