docker-mysql
============

[![Build Status](https://travis-ci.org/nanofi/docker-mysql.svg?branch=master)](https://travis-ci.org/nanofi/docker-mysql)

A Dockerfile that installs a mysql server with multiple databases. databases and users will generate dynamically with starting or stopping or killing a container.


## Usage
To run it:
```
$ docker run -d -v /var/run/docker.sock:/var/run/docker.sock -e MYSQL_ROOT_PASSWORD="password" nanofi/mysql
```
Start any containers with env vars:
- `MYSQL_DB_NAME`; database name
- `MYSQL_DB_USER`; username 
- `MYSQL_DB_PASS`; password

If `MYSQL_DB_NAME` presences, the container creates a database. If `MYSQL_DB_NAME`, `MYSQL_DB_USER` and `MYSQL_DB_PASS` presence, the container creates an user which has all privilege to the database.

