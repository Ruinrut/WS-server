FROM golang:1.16-alpine AS builder
RUN mkdir /app
ADD . /app
WORKDIR /app
RUN go build -o main .

FROM alpine AS final
COPY --from=builder /app/main /app/
EXPOSE 8080
ENTRYPOINT ["/app/main"]
