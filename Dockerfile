FROM golang:1.22-bookworm AS builder

WORKDIR /app

COPY main.go ./

RUN CGO_ENABLED=0 GOOS=linux go build -o /manager main.go


FROM debian:bookworm-slim

WORKDIR /

COPY --from=builder /manager /manager

EXPOSE 3000

ENTRYPOINT ["/manager"]
