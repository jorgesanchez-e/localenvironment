# Etapa de compilación
FROM golang:1.26-alpine AS builder

ARG GOOS
ARG GOARCH
ARG APPVERSION
ARG APPNAME

ENV GOOS=${GOOS}
ENV GOARCH=${GOARCH}
ENV APPVERSION=${APPVERSION}
ENV APPNAME=${APPNAME}

WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=$GOOS GOARCH=$GOARCH go build -o ${APPNAME} ./cmd/main.go

# Etapa final (mínima)
FROM scratch
COPY --from=builder /app/ddns /ddns
ENTRYPOINT ["/ddns"]