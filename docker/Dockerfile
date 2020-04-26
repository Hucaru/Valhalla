FROM golang:buster

WORKDIR /app
COPY . /app

RUN go get
RUN go build
RUN chmod +x Valhalla

RUN echo "#!/bin/sh" > docker-entrypoint.sh && \
  echo "set -ex" >> docker-entrypoint.sh && \
  echo 'exec "$@"' >> docker-entrypoint.sh && \
  chmod +x docker-entrypoint.sh

EXPOSE 8484 8585 8584 8685

ENTRYPOINT ["/app/docker-entrypoint.sh"]
CMD ./Valhalla -type login -config ./config_login.toml