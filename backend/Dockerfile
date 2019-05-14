FROM golang:latest as builder

COPY . /backend
WORKDIR /backend
ENV GO111MODULE=on
RUN CGO_ENABLED=0 GOOOS=linux go build -o backend

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /backend .
CMD ["./backend"]
