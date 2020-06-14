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

WORKDIR /go/src/github.com/ottotech/godic

RUN ["chmod", "+x", "/go/src/github.com/ottotech/godic/docker-entrypoint.sh"]
CMD ["/go/src/github.com/ottotech/godic/docker-entrypoint.sh"]

# Expose default http server port.
EXPOSE 8080
