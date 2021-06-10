FROM golang:1.16-alpine

WORKDIR /go/src/app
COPY ./src ./

ENV GOPROXY=https://goproxy.cn

RUN apk --no-cache add tzdata  && \
    ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    echo "Asia/Shanghai" > /etc/timezone

RUN go get -d -v ./...
RUN go install -v ./...

CMD ["chia-block-sync", "run"]
