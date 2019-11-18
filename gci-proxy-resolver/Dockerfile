FROM openfaas/of-watchdog:0.7.2 as watchdog

FROM golang:1.11-alpine3.10 as builder

# Allows you to add additional packages via build-arg
ARG ADDITIONAL_PACKAGE
ARG CGO_ENABLED=0

COPY --from=watchdog /fwatchdog /usr/bin/fwatchdog
RUN chmod +x /usr/bin/fwatchdog

WORKDIR /go/src/handler
COPY . .

# Run a gofmt and exclude all vendored code.
RUN test -z "$(gofmt -l $(find . -type f -name '*.go' -not -path "./vendor/*" -not -path "./function/vendor/*"))" || { echo "Run \"gofmt -s -w\" on your Golang code"; exit 1; }

RUN CGO_ENABLED=${CGO_ENABLED} GOOS=linux \
    go build --ldflags "-s -w" -a -installsuffix cgo -o gci-proxy-resolver . && \
    go test $(go list ./... | grep -v /vendor/) -cover

FROM alpine:3.10
RUN apk --no-cache add \
    ca-certificates

# Add non root user
RUN addgroup -S app && adduser -S -g app app
RUN mkdir -p /home/app

WORKDIR /home/app

COPY --from=builder /usr/bin/fwatchdog         .
COPY --from=builder /go/src/handler/handler    .

RUN chown -R app /home/app

USER app

ENV upstream_url="http://127.0.0.1:8082"
ENV mode="http"

ENV fprocess="./gci-proxy-resolver"
ENV MANAGER_ADDRESSES="150.165.75.121:5555,150.165.75.125:5555"
EXPOSE 8080

HEALTHCHECK --interval=3s CMD [ -e /tmp/.lock ] || exit 1

CMD ["./fwatchdog"]
