FROM golang:1.20 as BUILDER

MAINTAINER wanghao75<shalldows@163.com>

# build binary
WORKDIR /go/src/github.com/wanghao75/robot-invoke-jenkins
COPY . .
RUN GO111MODULE=on CGO_ENABLED=0 go build -a -o robot-invoke-jenkins .

# copy binary config and utils
FROM alpine:3.14
COPY  --from=BUILDER /go/src/github.com/wanghao75/robot-invoke-jenkins/robot-invoke-jenkins /opt/app/robot-invoke-jenkins

ENTRYPOINT ["/opt/app/robot-invoke-jenkins"]