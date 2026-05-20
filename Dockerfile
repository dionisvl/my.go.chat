ARG GO_VERSION=1.26.2
ARG ALPINE_VERSION=3.23

FROM golang:${GO_VERSION}-alpine${ALPINE_VERSION} AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
# Trim unused goose DB drivers (sqlite, mysql, etc.) — we only target postgres.
ARG GOOSE_TAGS="no_sqlite3 no_clickhouse no_mssql no_mysql no_vertica no_ydb no_libsql no_duckdb"
ARG APP_VERSION=dev
RUN CGO_ENABLED=0 GOOS=linux go build \
    -tags="${GOOSE_TAGS}" \
    -ldflags="-X 'mygochat/internal/config.Version=${APP_VERSION}'" \
    -o /bin/mygochat ./cmd/server

FROM alpine:${ALPINE_VERSION} AS runtime
RUN apk add --no-cache ca-certificates wget \
    && addgroup -S app && adduser -S app -G app
COPY --from=builder /bin/mygochat /bin/mygochat
USER app
EXPOSE 8080
HEALTHCHECK --interval=10s --timeout=3s --retries=5 \
    CMD wget --quiet --tries=1 --spider http://localhost:8080/health || exit 1
ENTRYPOINT ["/bin/mygochat"]
