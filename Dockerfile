FROM node:24-alpine AS fe_builder

WORKDIR /app/fe

RUN npm install --global pnpm@10.8.1

COPY htqrcode/fe/package.json htqrcode/fe/pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile

COPY htqrcode/fe/angular.json htqrcode/fe/tsconfig*.json ./
COPY htqrcode/fe/src/ ./src/
COPY htqrcode/fe/public/ ./public/

RUN pnpm run build --configuration production

FROM golang:1.27.1-alpine AS builder

RUN apk update && \
    apk upgrade -U && \
    apk --no-cache add ca-certificates pkgconf gcc libc-dev musl-dev git && \
    update-ca-certificates && \
    git config --global http.sslVerify false

RUN mkdir /app

WORKDIR /app

COPY htqrcode/go.mod htqrcode/go.sum ./

COPY htqrcode/ /app

COPY --from=fe_builder /app/client/api/ /app/client/api/

RUN rm -rf ~/.cache/go-build

RUN CGO_ENABLED=1 GOPROXY=direct,off GOINSECURE=* go build -tags musl -o /app/htqrcode /app/cmd

RUN chmod +x /app/htqrcode

## Build an image
FROM golang:1.27.1-alpine

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /app/htqrcode /app

CMD [ "/app/htqrcode" ]
