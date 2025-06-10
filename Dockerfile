FROM golang:1.23.4 AS builder
WORKDIR /app 
COPY go.* ./
RUN go mod download && go mod verify
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o observe_api ./cmd/api

FROM scratch
WORKDIR /app
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/observe_api .
COPY --from=builder /app/doc ./
EXPOSE 8080
CMD ["./observe_api"]
