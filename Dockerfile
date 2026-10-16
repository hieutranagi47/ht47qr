# Build the server with the checked-in client/api and Caro web assets.
# Generated Go sources and web assets must be present before building.
FROM golang:1.27.1-alpine AS builder

# Trust the corporate CA before downloading OS packages or Go modules.
COPY corp-ca.crt /usr/local/share/ca-certificates/corp-ca.crt
RUN cat /usr/local/share/ca-certificates/corp-ca.crt >> /etc/ssl/certs/ca-certificates.crt \
    && apk add --no-cache ca-certificates gcc musl-dev \
    && update-ca-certificates

WORKDIR /app

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download

COPY . ./
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=1 go build -mod=readonly -tags musl -trimpath -o /out/htqrcode ./cmd

FROM alpine:3.22 AS runtime

COPY corp-ca.crt /usr/local/share/ca-certificates/corp-ca.crt
RUN cat /usr/local/share/ca-certificates/corp-ca.crt >> /etc/ssl/certs/ca-certificates.crt \
    && apk add --no-cache ca-certificates tzdata \
    && update-ca-certificates \
    && addgroup -S app \
    && adduser -S -G app app

WORKDIR /app
RUN mkdir -p /app/data && chown app:app /app/data
COPY --from=builder /out/htqrcode ./htqrcode

USER app

# Supply DATABASE_URL and TLS_CERT/TLS_KEY at runtime (Compose does this).
ENV SERVER_PORT=8080 \
    SERVER_PORT_TLS=8443 \
    SERVER_SSE_PORT=8886 \
    SERVER_SSE_PORT_TLS=8446 \
    SERVER_GRPC_PORT="" \
    DATABASE_BACKEND=postgres \
    SQLITE_PATH=/app/data/htqrcode.db \
    TZ=Asia/Ho_Chi_Minh

EXPOSE 8080 8443 8886 8446

HEALTHCHECK --interval=30s --timeout=3s --start-period=40s --retries=3 \
    CMD wget -q -O /dev/null "http://127.0.0.1:${SERVER_PORT}/health" || exit 1

ENTRYPOINT ["/app/htqrcode"]
