FROM golang:1.21 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /balancer ./cmd/balancer

FROM gcr.io/distroless/static:nonroot

COPY --from=builder /balancer /balancer
COPY config/config.json /config/config.json

ENTRYPOINT ["/balancer"]
