FROM golang:1.20 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN pwd
RUN ls -la
RUN go mod download
COPY *.go ./
RUN mkdir /app/scrap_data
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o reviews_scrap_redis

FROM nvidia/cuda:11.6.2-base-ubuntu20.04
WORKDIR /app
COPY --from=builder /app/reviews_scrap_redis /app/reviews_scrap_redis
COPY --from=builder /app/scrap_data /app/scrap_data
ENTRYPOINT ["./reviews_scrap_redis"]
