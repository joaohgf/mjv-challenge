FROM golang:1.26-alpine AS build

WORKDIR /src
COPY go.mod go.sum* ./
RUN go mod download
COPY . .

ARG APP
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/app ./cmd/${APP}

FROM alpine:3.22
RUN adduser -D -H appuser
USER appuser
COPY --from=build /out/app /app
ENTRYPOINT ["/app"]
