FROM golang:1.26 AS builder

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o service_a .

FROM scratch

COPY --from=builder /app/service_a /service_a

EXPOSE 3333

ENTRYPOINT ["/service_a"]