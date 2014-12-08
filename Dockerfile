FROM mysql:5.7

MAINTAINER nanofi <nanogenomu@gmail.com>

RUN usermod -u 1000 mysql
COPY docker-entrypoint.sh /entrypoint.sh

CMD ["mysqld_safe", "--datadir=/var/lib/mysql", "--user=mysql"]
