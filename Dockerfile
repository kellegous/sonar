FROM kellegous/build:0d98364e AS build

COPY . /src

RUN cd /src && make nuke ALL

FROM lsiobase/debian:bookworm

RUN apt-get update \
  && apt-get install -y ca-certificates tzdata jq iptables \
  && apt-get clean

COPY --from=build /src/bin/sonard /usr/local/bin/sonard

EXPOSE 8080

CMD ["/usr/bin/with-contenv", "/usr/local/bin/sonard", "--conf=/data/sonar.toml"]
