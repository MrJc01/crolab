FROM golang:1.23-alpine AS builder
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o crolab ./cmd/crolab/

FROM alpine:3.19
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /build/crolab .
COPY --from=builder /build/web ./web
EXPOSE 8844 8855
VOLUME ["/app/data"]
ENTRYPOINT ["./crolab"]
CMD ["provider", "--admin-port", ":8844", "--client-port", ":8855", "--db", "/app/data/crolab.db"]
