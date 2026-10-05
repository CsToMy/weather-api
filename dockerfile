# syntax=docker/dockerfile:1

ARG GO_VERSION=1.27

FROM golang:${GO_VERSION} AS build
WORKDIR /src

COPY go.mod go.sum* ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/weather-api ./cmd/server

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/weather-api /weather-api
EXPOSE 8080
ENTRYPOINT ["/weather-api"]