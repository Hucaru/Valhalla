FROM iron/go:dev

WORKDIR /app

ENV SRC_DIR=/go/src/github.com/Hucaru/Valhalla/

RUN go get -u github.com/go-sql-driver/mysql

ADD . $SRC_DIR

RUN cd $SRC_DIR; go build -o app; cp app /app/

ENTRYPOINT ["./app"]