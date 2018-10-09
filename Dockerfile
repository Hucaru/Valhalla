FROM iron/go:dev

WORKDIR /app

ENV SRC_DIR=/go/src/github.com/Hucaru/Valhalla/

RUN go get -u github.com/go-sql-driver/mysql
RUN go get -u golang.org/x/sys/...
RUN go get -u github.com/fsnotify/fsnotify 
RUN go get -u github.com/mattn/anko
RUN go get -u github.com/google/uuid
RUN go get -u github.com/BurntSushi/toml

ADD . $SRC_DIR

RUN cd $SRC_DIR; go build -o app; cp app /app/

ENTRYPOINT ["./app"]