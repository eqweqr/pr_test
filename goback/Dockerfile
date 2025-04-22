FROM golang:latest as builder
RUN mkdir /app
COPY ./raft/ /app/
WORKDIR /app
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

FROM alpine:latest
COPY --from=builder /app/main /home/main
RUN chmod 555 /home/main
ENTRYPOINT ["/home/main"]
