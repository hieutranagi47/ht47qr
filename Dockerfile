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
    TZ=Asia/Ho_Chi_Minh \
    # Mặc định, Go sẽ lấy số CPU của máy Host làm giá trị GOMAXPROCS. Nếu bạn chạy container trên một server 64 cores nhưng chỉ giới hạn tài nguyên container là cpus: 2, Go (từ phiên bản 1.24 trở về trước) vẫn có thể hiểu nhầm là có 64 cores, dẫn đến việc tranh chấp CPU và giảm hiệu năng (CPU throttling).
    GOMAXPROCS="2" \
    # =In ra trạng thái của bộ lập lịch (Scheduler) sau mỗi X mili-giây, giúp bạn xem các Goroutine trong Echo đang phân bổ trên các luồng CPU như thế nào để debug hiện tượng nghẽn.
    GODEBUG="madvdontneed=1,netdns=go,gctrace=1,schedtrace=1000" \
    # ~ 90% of the RAM. Đặt giới hạn mềm cho tổng lượng bộ nhớ mà Go Runtime được phép dùng. Khi bộ nhớ chạm ngưỡng này, bộ thu gom rác (Garbage Collector - GC) sẽ bị kích hoạt liên tục và quyết liệt hơn để giải phóng RAM, cố gắng không để container vượt quá giới hạn và bị Docker Kill.
    GOMEMLIMIT="900MiB" \
    # Xác định tỷ lệ phần trăm tăng trưởng của bộ nhớ heap trước khi đợt thu gom rác tiếp theo được kích hoạt (mặc định là 100).
    GOGC="90" \
    # Force TURN locally so a successful game proves the relay works.
    CARO_TURN_URLS="turn:free.expressturn.com:3478?transport=tcp,turn:free.expressturn.com:3478?transport=udp" \
    CARO_TURN_USERNAME="000000002107744556" \
    CARO_TURN_PASSWORD="mxJLG/ft6aioH97b+m7BNYTlVv4=" \
    CARO_ICE_TRANSPORT_POLICY="${CARO_ICE_TRANSPORT_POLICY:-relay}" \
    # Container listeners are fixed; SERVER_PORT below selects the host port.
    SERVER_PORT="8080" \
    DATABASE_BACKEND=postgres \
    # Use the Compose service name, not localhost, to reach PostgreSQL.
    # Use URL-safe values for these connection-string components.
    DATABASE_URL="postgresql://${POSTGRES_USER:-htqrcode}:${POSTGRES_PASSWORD:-htqrcode-local}@postgres:5432/${POSTGRES_DB:-htqrcode}?sslmode=disable" \
    DATABASE_URL_UNPOOLED="" \
    POSTGRES_URL="" \
    POSTGRES_URL_NON_POOLING="" \
    SERVER_GRPC_PORT="" \
    SERVER_PORT_TLS="8443" \
    SERVER_SSE_PORT="8886" \
    SERVER_SSE_PORT_TLS="8446" \
    TLS_CERT="-----BEGIN CERTIFICATE-----\\nMIIEVjCCAr6gAwIBAgIQBLHLN215TlB91k7qD1nqxTANBgkqhkiG9w0BAQsFADCB\\nhTEeMBwGA1UEChMVbWtjZXJ0IGRldmVsb3BtZW50IENBMS0wKwYDVQQLDCRoaWV1\\ndHJhbkBIaWV1cy1NQlAubG9jYWwgKEhpZXUgVHJhbikxNDAyBgNVBAMMK21rY2Vy\\ndCBoaWV1dHJhbkBIaWV1cy1NQlAubG9jYWwgKEhpZXUgVHJhbikwHhcNMjUxMTIy\\nMDgzMDQwWhcNMjgwMjIyMDgzMDQwWjBYMScwJQYDVQQKEx5ta2NlcnQgZGV2ZWxv\\ncG1lbnQgY2VydGlmaWNhdGUxLTArBgNVBAsMJGhpZXV0cmFuQEhpZXVzLU1CUC5s\\nb2NhbCAoSGlldSBUcmFuKTCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEB\\nALvWrAZXyGPRdFLGpm5N5w0yYdw49FKgLZsHM+TfQd68iChfdarVPKxusQ652ELU\\n3yKuZgvtyWQFQ+3J2FhyWyFoz5fbdkSM0BviHEqwb6/goD6oDRCxUlOWxQJwlAJg\\npM6O17kVZIujO56wCdRFLSbWXal5oiDni1Y2+FMyH3e1qvctWOeOkfIe6enxbbGz\\nVyjthwSrt3gxhlL51k0VyVvzIhWkw3cFzbBv3sgZHvfrI//QxYsvjAEDLv0aTUTc\\nqUOYr4Ukh0p1JHf3AjHKZB0JSxVJu+NxBm9Tv7GgKoonKFQOwFmuIwtQMzjw4C5I\\nlPGb0uXYDT5VmbPBfJHDPIsCAwEAAaNuMGwwDgYDVR0PAQH/BAQDAgWgMBMGA1Ud\\nJQQMMAoGCCsGAQUFBwMBMB8GA1UdIwQYMBaAFKVMT13NSG8ZXwJiyFrjt71exNVz\\nMCQGA1UdEQQdMBuCCWxvY2FsaG9zdIIObG9jYWwuaHQ0Ny5jb20wDQYJKoZIhvcN\\nAQELBQADggGBAIFgHGax+DcMHzPDFVB4SD2MD+b7340hZwKKpwvu/raYfOg4C47M\\nbEhefzdvIUcNMt/ra4uRAeyZhBDhC//DvOMkUjuMv3RDkHfRI+r+63Y4xiCe076j\\n1b8hLYIVoyuY9RKZMGXsVxi1d/2YXCiNgWTLfSSCeuewcFxUwzqbkQI8fY0DKLTL\\nw3X4g4ifiIRrDWc90O/rs0HMBmPQxTQgR9g4lS5fP8otIPadTnVKR4mKpr8BeBtv\\nyFKnUWEAFZxM7ix1y5a4p421KNAu9L7roJmyDbqWlBRO3H7+DjJBYfj4bED+Xz5p\\nCVkdBxYfq0r/ZxE4+NMbtVfeBq+7ctQEzEehIe2OEfH1qNVQTT3poMM5kbvHcR6U\\n37qmex01rHFQImRbvuX3qThwQkU3Z3UH0eJdeXgXypTTFFVtyEwgGeDc8l7odkGc\\nTdgviMmi6MfGBFb1sU1nnvg91+xxkRUMhw6E78RXk616PmsndeVKqBGingjsxMZt\\nJ4e232r9YihAfw==\\n-----END CERTIFICATE-----" \
    TLS_KEY="-----BEGIN PRIVATE KEY-----\\nMIIEwAIBADANBgkqhkiG9w0BAQEFAASCBKowggSmAgEAAoIBAQC71qwGV8hj0XRS\\nxqZuTecNMmHcOPRSoC2bBzPk30HevIgoX3Wq1TysbrEOudhC1N8irmYL7clkBUPt\\nydhYclshaM+X23ZEjNAb4hxKsG+v4KA+qA0QsVJTlsUCcJQCYKTOjte5FWSLozue\\nsAnURS0m1l2peaIg54tWNvhTMh93tar3LVjnjpHyHunp8W2xs1co7YcEq7d4MYZS\\n+dZNFclb8yIVpMN3Bc2wb97IGR736yP/0MWLL4wBAy79Gk1E3KlDmK+FJIdKdSR3\\n9wIxymQdCUsVSbvjcQZvU7+xoCqKJyhUDsBZriMLUDM48OAuSJTxm9Ll2A0+VZmz\\nwXyRwzyLAgMBAAECggEBAInj51iorFSKeCgj6B1lVs4bRDeUH9df+tqpiPDk68gF\\nArcQjKKh8u5qO9QhnZ5CAqPE/uiW4JZqPxmaaRVUrY1zEaHa25ErBcLSVA6IQxUL\\ngtc8KossgNEYBvvphN3R8Gop9jbKe4r+g0/UzwJi3eh6yTsWLWyWTXU22UjJrMwE\\n7sdQggo/sZkj/MtnYeVDJRhEn/jXjDLsG+nXn7o0XMxpl52lZZd2shKezLatFtUD\\nyUIZ1eSjrbn1HrB1qmM7jlxfoct3SK61/xirLCFlh1Fo0silz6DvuSQlX2DKXYTe\\nCceqx1iz0v0iY197bp7k6pZsziob3CRVKLCvUFjKroECgYEAxAkGlXXMtIer3OnK\\n6G2K/kcJQV93tLd9ETV51Hxb5q0yGL+ffTY7/CnxEehH+GIoX59lYcoTBhDuRktb\\nvG6+/kh6NhRDaxK4OwsduJcI5PjN/INWU+3eLbgj8XfMfN5jgwpfEsIQnSBVnfLS\\nKhUrATCoiEuXPl0kW4Omm10eh+ECgYEA9UvJNetYwziB5ajJfVAioWF2UJPpejL9\\nnmLhJu+/dOn/Xj6BpOai1phArJlbA27ZbwazAyFtIvH+Jv8XMK26jvtB5tEBWwSQ\\n9dHlUdavMCkCYV4Ar6m4CngLGlat8HHXNp7PlAhB8q1HOeDzP2nnTea70hsWpF+V\\ncfai/SxIoesCgYEAuyX6uRka8oZLVr2Dzt01QofCmDxx5X7+0T0CMA8BFXnnR6U4\\nt6rdHWFVGcltW/0pmhBsEy19d98V+tnanEwF3nCG67XwnvniE8wK9byK6z10OUuR\\nUNLXt0ICXnYGLHeK/QEbvqayga7qzb2RUkefJKW1X1j67LizCwLmPZBqo0ECgYEA\\nppdSztDIVsVgyvWH6hhrfIixjuX1ki1hO5lR6ly2/ZMSARKYf8K3YuDCQ7ZfLoSx\\nHQtCtmSUN0B8aEWknUXPxt1zk2vs7MC8JobnpF4WWrHvUOtDsgmk2U9P0xJ9TSgz\\njn/q4FF/YG44SmGxMicdlTWX54jIJthXqa9kYGcb6QMCgYEAngJpPCVS2e1CLl3+\\nm6S8DO4GmK0JTvIPu7rh02mkwbw0SSqtmhqIGVs2D16XPzXkVhBG29CJDV5/vrmo\\nVKc/d9tCT1PeqvZAxHxjZr+RMTuDNn73xYTAuDXDfqWlAGpeWDwFJAevSZioUP8S\\n9h1+gR88VJwNpZjX34Z3Ff7oZLE=\\n-----END PRIVATE KEY-----"


EXPOSE 8080 8443 8886 8446

HEALTHCHECK --interval=30s --timeout=3s --start-period=40s --retries=3 \
    CMD wget -q -O /dev/null "http://127.0.0.1:${SERVER_PORT}/health" || exit 1

ENTRYPOINT ["/app/htqrcode"]
