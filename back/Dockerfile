FROM golang:1.24-alpine AS base
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .

###########################################

FROM base AS dev
RUN go install github.com/air-verse/air@v1.62.0
RUN mkdir -p /air/bin /air/logs /air/tmp /air/cache
EXPOSE 8080
COPY dev.air.toml /air/air.toml
CMD ["air", "-c", "/air/air.toml"]

###########################################

FROM base AS builder
ARG VERSION="dev"
ARG COMMIT="unknown"
ARG BUILT="unknown"

ENV CGO_ENABLED=0 GOOS=linux

RUN LDFLAGS="-s -w -extldflags=-static \
    -X main.version=${VERSION} \
    -X main.commit=${COMMIT} \
    -X main.built=${BUILT}" \
    && go build -trimpath -ldflags="$LDFLAGS" -o /out/app .

###########################################

FROM alpine:3 AS certs
RUN apk add --no-cache ca-certificates

###########################################

FROM scratch AS prod
COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
ARG APP_NAME=app
COPY --from=builder /out/app /app
USER 65532:65532
EXPOSE 8080
ENTRYPOINT ["/app"]
