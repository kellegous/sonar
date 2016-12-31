FROM ubuntu:xenial

COPY bin/sonard /usr/local/bin/sonard

RUN apt-get update \
  && apt-get install -y ca-certificates \
  && apt-get clean

EXPOSE 7676

CMD ["/usr/local/bin/sonard", "--data=/data"]
