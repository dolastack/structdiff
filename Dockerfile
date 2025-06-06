# Dockerfile
FROM golang:1.21 as builder
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 go build -o /structdiff

FROM gcr.io/distroless/static-debian12
COPY --from=builder /structdiff /usr/local/bin/
ENTRYPOINT ["/usr/local/bin/structdiff"]