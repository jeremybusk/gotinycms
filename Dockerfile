# syntax=docker/dockerfile:1
FROM node:22-alpine AS web
WORKDIR /src/web
COPY web/package.json web/package-lock.json* ./
RUN npm ci
COPY web ./
RUN npm run build

FROM golang:1.26-alpine AS go
RUN apk add --no-cache build-base git
WORKDIR /src
COPY go.mod go.sum* ./
RUN GOPROXY=direct go mod download
COPY . .
COPY --from=web /src/web/dist ./web/dist
RUN CGO_ENABLED=1 go build -trimpath -ldflags='-s -w' -o /out/uvoocms ./cmd/uvoocms

FROM alpine:3.21
RUN apk add --no-cache ca-certificates sqlite-libs && adduser -D -H -u 10001 uvoocms
WORKDIR /app
COPY --from=go /out/uvoocms /app/uvoocms
COPY --from=web /src/web/dist /app/web/dist
RUN mkdir -p /data/uploads && chown -R uvoocms:uvoocms /data /app
USER uvoocms
ENV CMS_ADDR=:8080 CMS_DATA_DIR=/data CMS_DB=/data/cms.db CMS_UPLOAD_DIR=/data/uploads
EXPOSE 8080
VOLUME ["/data"]
ENTRYPOINT ["/app/uvoocms"]
