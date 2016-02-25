# datafactory

FROM golang:1.5
MAINTAINER Zonesan <chaizs@asiainfo.com>

ENV TIME_ZONE=Asia/Shanghai
RUN ln -snf /usr/share/zoneinfo/$TIME_ZONE /etc/localtime && echo $TIME_ZONE > /etc/timezone

WORKDIR datafactory

ADD . .

RUN make build && \
    mv _output/local/bin/linux/amd64/openshift /bin/ && \
    rm -rf ../datafactory /usr/local/go

WORKDIR /root

EXPOSE 8443 4001 53 10250 7001

ENTRYPOINT ["openshift", "start"]


