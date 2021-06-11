FROM golang:1.16-alpine as build_stage

WORKDIR /go/src/app
COPY ./src ./

ENV GOPROXY=https://goproxy.cn

RUN go get -d -v ./...
RUN go build -v ./...

FROM alpine:3.14

RUN apk --no-cache add tzdata  && \
    ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    echo "Asia/Shanghai" > /etc/timezone

WORKDIR /go/src/app

COPY --from=build_stage /go/src/app/chia-reporter /usr/bin/

CMD ["chia-reporter", "export"]
