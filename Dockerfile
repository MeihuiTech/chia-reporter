FROM golang:1.16-alpine

WORKDIR /go/src/app
COPY ./src ./

ENV GOPROXY=https://goproxy.cn
RUN go get -d -v ./...
RUN go install -v ./...

CMD ["chia-block-sync", "run"]
