# syntax=docker/dockerfile:1

ARG GO_VERSION=1.26.2
ARG CMD=server

# ---------- Build stage ----------
FROM golang:${GO_VERSION}-alpine AS builder

ARG CMD

WORKDIR /src

RUN apk add --no-cache ca-certificates upx

ENV CGO_ENABLED=0

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN mkdir -p /out \
    && go build -trimpath -ldflags="-s -w" -o /out/calendar-manager ./cmd/${CMD}

RUN upx --lzma --best /out/calendar-manager || true

# ---------- Dev stage ----------
FROM golang:${GO_VERSION}-alpine AS dev
COPY --from=builder /out/calendar-manager /app/calendar-manager
ENTRYPOINT ["/app/calendar-manager"]

# ---------- Runtime stage ----------
FROM gcr.io/distroless/static:nonroot AS runner

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

WORKDIR /app
COPY --from=builder /out/calendar-manager /app/calendar-manager

USER nonroot:nonroot
ENTRYPOINT ["/app/calendar-manager"]
