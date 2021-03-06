FROM golang:alpine

ENV PRODUCTION="true"

RUN apk --no-cache add --virtual build-dependencies \
    git \
    && apk update && apk upgrade && apk add bash \
    && go get -d github.com/ottotech/godic \
    && cd $GOPATH/src/github.com/ottotech/godic \
    && go get github.com/jteeuwen/go-bindata/... \
    && go-bindata assets/ \
    && go build .

# Forward error logs to docker log collector
RUN touch /go/src/github.com/ottotech/godic/error.log \
    && ln -sf /dev/stderr /go/src/github.com/ottotech/godic/error.log

ADD docker-entrypoint.sh /go/src/github.com/ottotech/godic/

VOLUME /go/src/github.com/ottotech/godic/data

WORKDIR /go/src/github.com/ottotech/godic

# Expose default http server port.
EXPOSE 8080

RUN ["chmod", "+x", "/go/src/github.com/ottotech/godic/docker-entrypoint.sh"]
CMD ["/go/src/github.com/ottotech/godic/docker-entrypoint.sh"]
