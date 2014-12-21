docker-mysql
============

[![Build Status](https://travis-ci.org/nanofi/docker-mysql.svg?branch=master)](https://travis-ci.org/nanofi/docker-mysql)

A Dockerfile that installs a mysql server with multiple databases. databases and users will generate dynamically with starting or stopping or killing a container.


## Usage
To run it:
```
$ docker run -d -v /var/run/docker.sock:/var/run/docker.sock nanofi/mysql
```
Start any containers with env vars:
- `DB_NAME`; database name
- `DB_USER`; username 
- `DB_PASS`; password

If `DB_NAME` presences, the container creates a database. If `DB_NAME`, `DB_USER` and `DB_PASS` presence, the container creates an user which has all privilege to the database.

