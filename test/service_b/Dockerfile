FROM golang:1.26 AS builder

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o service_b .


FROM scratch

COPY --from=builder /app/service_b /service_b


EXPOSE 4444

ENTRYPOINT ["/service_b"]