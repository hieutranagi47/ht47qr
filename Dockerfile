# Build the Angular frontend into client/api, which is embedded by the Go app.
FROM node:24-alpine AS frontend

WORKDIR /app/fe

RUN npm install --global pnpm@10.8.1

COPY fe/package.json fe/pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile

COPY fe/angular.json fe/tsconfig*.json ./
COPY fe/src/ ./src/
COPY fe/public/ ./public/

RUN pnpm run build --configuration development

# Build the Go server with the generated frontend assets.
FROM golang:1.27.1-alpine AS builder

RUN apk add --no-cache ca-certificates gcc libc-dev musl-dev protobuf

# Install the tools required by task gen (OpenAPI, protobuf, and formatting).
RUN --mount=type=cache,target=/go/pkg/mod \
    go install github.com/go-task/task/v3/cmd/task@v3.53.1 \
    && go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.12 \
    && go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.5.1

WORKDIR /app

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download

COPY . ./
COPY --from=frontend /app/client/api/ ./client/api/
RUN --mount=type=cache,target=/go/pkg/mod task gen

RUN CGO_ENABLED=1 go build -tags musl -trimpath -o /out/htqrcode ./cmd

# Run as an unprivileged user in a small runtime image.
FROM alpine:3.22 AS runtime

RUN apk add --no-cache ca-certificates \
    && addgroup -S app \
    && adduser -S -G app app

WORKDIR /app
RUN mkdir -p /app/data && chown app:app /app/data
COPY --from=builder /out/htqrcode ./htqrcode

USER app

ENV SERVER_PORT=8080 \
    SERVER_PORT_TLS=8443 \
    SERVER_SSE_PORT=8081 \
    SERVER_SSE_PORT_TLS=8444 \
    SQLITE_PATH=/app/data/htqrcode.db

EXPOSE 8080 8081 8443 8444
VOLUME ["/app/data"]

ENTRYPOINT ["/app/htqrcode"]
