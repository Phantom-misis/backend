FROM golang:1.25.4

RUN go mod init backend && go mod tidy && go build

EXPOSE 8080

CMD [./backend]