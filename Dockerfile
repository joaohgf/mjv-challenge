# syntax=docker/dockerfile:1.7
# Builds one of the Go entry points with cached Go modules.
FROM golang:1.26-alpine AS build

WORKDIR /src
COPY go.mod go.sum* ./
RUN --mount=type=cache,target=/go/pkg/mod,sharing=locked go mod download
COPY . .

ARG APP
RUN --mount=type=cache,target=/root/.cache/go-build,sharing=locked \
    --mount=type=cache,target=/go/pkg/mod,sharing=locked \
    CGO_ENABLED=0 GOOS=linux go build -o /out/app ./cmd/${APP}

FROM alpine:3.22
# The runtime image contains only the static binary and a non-root user.
RUN adduser -D -H appuser
USER appuser
COPY --from=build /out/app /app
ENTRYPOINT ["/app"]
