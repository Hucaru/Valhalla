FROM iron/go:dev

ENV SRC_DIR=/app

WORKDIR $SRC_DIR

ADD . $SRC_DIR

RUN cd $SRC_DIR; go build -o server

ENTRYPOINT ["/app/server"]