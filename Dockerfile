FROM golang:1.20-alpine AS builder
WORKDIR /app/service_go_wb
COPY go.mod ./
COPY go.sum ./
RUN go mod download
COPY *.go ./
RUN go build -o main .
RUN rm main.go

FROM alpine:3.18
WORKDIR /app/service_go_wb
COPY --from=builder /app/service_go_wb/ .
CMD [ "./main" ]