FROM kellegous/build:f1799259 AS build

ARG SHA
ARG BUILD_TIME

COPY . /src

RUN cd /src && make SHA=${SHA} BUILD_TIME=${BUILD_TIME} nuke ALL

FROM lsiobase/debian:bookworm

RUN apt-get update \
  && apt-get install -y ca-certificates tzdata jq iptables \
  && apt-get clean

COPY --from=build /src/bin/sonard /usr/local/bin/sonard

EXPOSE 8080

CMD ["/usr/bin/with-contenv", "/usr/local/bin/sonard", "--conf=/data/sonar.toml"]
