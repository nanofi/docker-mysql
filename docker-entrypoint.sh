#!/bin/bash
set -e

if [ ! -d '/var/lib/mysql/mysql' -a "${1%_safe}" = 'mysqld' ]; then
	if [ -z "$MYSQL_ROOT_PASSWORD" ]; then
		echo >&2 'error: database is uninitialized and MYSQL_ROOT_PASSWORD not set'
		echo >&2 '  Did you forget to add -e MYSQL_ROOT_PASSWORD=... ?'
		exit 1
	fi
	
	mysql_install_db --user=mysql --datadir=/var/lib/mysql
	
	# These statements _must_ be on individual lines, and _must_ end with
	# semicolons (no line breaks or comments are permitted).
	# TODO proper SQL escaping on ALL the things D:
	TEMP_FILE='/tmp/mysql-first-time.sql'
	cat > "$TEMP_FILE" <<-EOSQL
		DELETE FROM mysql.user ;
		CREATE USER 'root'@'%' IDENTIFIED BY '${MYSQL_ROOT_PASSWORD}' ;
		GRANT ALL ON *.* TO 'root'@'%' WITH GRANT OPTION ;
		DROP DATABASE IF EXISTS test ;
	EOSQL

	if [ "$MYSQL_DATABASES" ]; then
    for DATABASE in $(echo $MYSQL_DATABASES | sed 's/,/ /g'); do
      echo "CREATE DATABASE IF NOT EXISTS $DATABASE ;" >> "$TEMP_FILE"
    done
	fi
	
	if [ "$MYSQL_USERS" ]; then
    for USER in $(echo $MYSQL_USERS | sed 's/,/ /g'); do
      ARRY=($(echo $USER | sed 's/:/ /g'))
      NAME=${ARRY[0]}
      PASS=${ARRY[1]}
      GRANTS="${ARRY[2]}"
		  echo "CREATE USER '$NAME'@'%' IDENTIFIED BY '$PASS' ;" >> "$TEMP_FILE"

      if [ "$GRANTS" ]; then
        for GRANT in $(echo $GRANTS | sed 's/\// /g'); do
          echo "GRANT ALL ON $GRANT.* TO '$NAME'@'%' ;" >> "$TEMP_FILE"
        done
      fi
    done
	fi
	
	echo 'FLUSH PRIVILEGES ;' >> "$TEMP_FILE"
	
	set -- "$@" --init-file="$TEMP_FILE"
fi

chown -R mysql:mysql /var/lib/mysql
exec "$@"
