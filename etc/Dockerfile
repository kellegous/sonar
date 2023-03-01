FROM kellegous/build:22f1124f as build

COPY . /src

RUN cd /src && make nuke ALL

FROM ubuntu:jammy

RUN apt-get update \
  && apt-get install -y ca-certificates tzdata \
  && apt-get clean

COPY /src/bin/sonard /usr/local/bin/sonard

CMD ["/usr/local/bin/sonard", "--conf=/data/sonar.toml"]
