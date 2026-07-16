# Build stage
FROM golang:1.26-alpine AS builder
RUN apk add --no-cache ca-certificates

ARG TARGETOS
ARG TARGETARCH
ARG GOOS
ARG GOARCH
ARG APPVERSION
ARG APPNAME

WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 \
    GOOS="${GOOS:-$TARGETOS}" \
    GOARCH="${GOARCH:-$TARGETARCH}" \
    go build -o ${APPNAME} ./cmd/main.go

# Final stage (minimal)
FROM scratch
ARG APPNAME
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /app/${APPNAME} /${APPNAME}
ENTRYPOINT ["/ddns"]
