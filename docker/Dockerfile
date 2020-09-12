FROM golang:1.15.2

WORKDIR /app
COPY . /app

RUN go get
RUN go build
RUN chmod +x Valhalla

RUN echo "#!/bin/sh" > docker-entrypoint.sh && \
  echo "set -ex" >> docker-entrypoint.sh && \
  echo 'exec "$@"' >> docker-entrypoint.sh && \
  chmod +x docker-entrypoint.sh

ENTRYPOINT ["/app/docker-entrypoint.sh"]
CMD ["/app/Valhalla", "-type", "login", "-config", "/app/config_login.toml"]