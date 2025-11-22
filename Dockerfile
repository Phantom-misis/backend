FROM golang:1.25.4 AS builder_backend
WORKDIR /backend
COPY . .

ENV GOEXPERIMENT=greenteagc


RUN go mod init backend && go mod tidy && go build

FROM debian:trixie-slim

WORKDIR /backend
COPY --from=builder_backend /backend .

EXPOSE 8080
CMD ["./backend"]