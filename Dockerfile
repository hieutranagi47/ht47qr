# Build the Angular frontend that the Go client module embeds.
FROM node:24-alpine AS fe_builder

WORKDIR /app/fe

RUN npm install --global pnpm@10.8.1

COPY fe/package.json fe/pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile

COPY fe/angular.json fe/tsconfig*.json ./
COPY fe/src/ ./src/
COPY fe/public/ ./public/

RUN pnpm run build --configuration production

FROM golang:1.27.1-alpine AS builder

RUN apk add --no-cache ca-certificates pkgconf gcc libc-dev musl-dev git protobuf

# Install the generators used by task gen.
RUN --mount=type=cache,target=/go/pkg/mod \
    go install github.com/go-task/task/v3/cmd/task@v3.53.1 \
    && go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.12 \
    && go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.5.1 \
    && go install github.com/sqlc-dev/sqlc/cmd/sqlc@v1.31.1

WORKDIR /app

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download

COPY . ./

# Use the production frontend in the Go embed directory, then regenerate Go sources.
RUN rm -rf /app/client/api
COPY --from=fe_builder /app/client/api/ /app/client/api/

RUN --mount=type=cache,target=/go/pkg/mod task gen
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=1 go build -mod=readonly -tags musl -trimpath -o /out/htqrcode ./cmd

FROM alpine:3.22 AS runtime

RUN apk add --no-cache ca-certificates tzdata \
    && addgroup -S app \
    && adduser -S -G app app

WORKDIR /app
RUN mkdir -p /app/data && chown app:app /app/data
COPY --from=builder /out/htqrcode ./htqrcode

USER app

# Supply DATABASE_URL at runtime. TLS listeners are disabled unless their ports
# and matching TLS_CERT/TLS_KEY values are configured by the cloud environment.
ENV SERVER_PORT=8080 \
    SERVER_PORT_TLS="" \
    SERVER_SSE_PORT=8886 \
    SERVER_SSE_PORT_TLS="" \
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
    CARO_TURN_URLS="" \
    CARO_TURN_USERNAME="" \
    CARO_TURN_PASSWORD="" \
    CARO_ICE_TRANSPORT_POLICY="all" \
    DATABASE_URL="" \
    DATABASE_URL_UNPOOLED="" \
    POSTGRES_URL="" \
    POSTGRES_URL_NON_POOLING="" \
    TLS_CERT="" \
    TLS_KEY=""


EXPOSE 8080 8443 8886 8446

HEALTHCHECK --interval=30s --timeout=3s --start-period=40s --retries=3 \
    CMD wget -q -O /dev/null "http://127.0.0.1:${SERVER_PORT}/health" || exit 1

ENTRYPOINT ["/app/htqrcode"]
