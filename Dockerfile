FROM mysql:5.7

MAINTAINER nanofi <nanogenomu@gmail.com>

RUN \
  apt-get update && \
  apt-get install -y --no-install-recommends ca-certificates wget && \
  apt-get clean && rm -rf /var/cache/apt/archives/* /var/lib/apt/lists/*

RUN \
  wget -P /usr/local/bin https://godist.herokuapp.com/projects/ddollar/forego/releases/current/linux-amd64/forego && \
  chmod u+x /usr/local/bin/forego

ENV CONTAINER_VERSION 0.2.0
RUN \
  wget https://github.com/nanofi/docker-mysql/releases/download/$CONTAINER_VERSION/docker-mysql-linux-amd64-$CONTAINER_VERSION.tar.gz && \
  tar -xvzf docker-mysql-linux-amd64-$CONTAINER_VERSION.tar.gz && \
  mv dist/docker-mysql-linux-amd64 /usr/local/bin/docker-mysql && \
  rm -rf dist

COPY . /app/
WORKDIR /app/

ENTRYPOINT []
CMD ["forego", "start", "-r"]
