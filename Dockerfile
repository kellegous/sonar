FROM kellegous/build:e12b9a31 as build

COPY . /src

RUN cd /src && make nuke ALL

FROM ubuntu:jammy

RUN apt-get update \
  && apt-get install -y ca-certificates tzdata \
  && apt-get clean

COPY --from=build /src/bin/sonard /usr/local/bin/sonard

CMD ["/usr/local/bin/sonard", "--conf=/data/sonar.toml"]
