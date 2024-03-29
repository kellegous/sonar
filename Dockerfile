FROM kellegous/build:0d98364e as build

COPY . /src

RUN cd /src && make nuke ALL

FROM ubuntu:jammy

RUN apt-get update \
  && apt-get install -y ca-certificates tzdata \
  && apt-get clean

COPY --from=build /src/bin/sonard /usr/local/bin/sonard

CMD ["/usr/local/bin/sonard", "--conf=/data/sonar.toml"]
