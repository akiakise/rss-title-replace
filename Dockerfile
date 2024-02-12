FROM golang:1.21

WORKDIR /app

COPY go.mod go.sum *.go ./

RUN go env -w GOPROXY=https://goproxy.cn,direct && go mod download

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/rss-title-replace

EXPOSE 8080

CMD ["/app/rss-title-replace"]