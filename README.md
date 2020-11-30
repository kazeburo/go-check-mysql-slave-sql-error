# go-check-mysql-slave-sql-error

Mackerel check plugin for MySQL replica/slave sql error

check-mysql-slave-sql-error will raize critical alert when Last_SQL_Error is not null.

## Usage

```
Usage:
  check-mysql-slave-sql-error [OPTIONS]

Application Options:
  -H, --host=     Hostname (default: localhost)
  -p, --port=     Port (default: 3306)
  -u, --user=     Username (default: root)
  -P, --password= Password
      --timeout=  Timeout to connect mysql (default: 10s)
  -v, --version   Show version

Help Options:
  -h, --help      Show this help message
```

Example

```
$ ./check-mysql-slave-sql-error --user=xxx --password=xxx
mysql-slave-sql-error - MySQL replica/slave SQL error CRITICAL: Last_SQL_Error found: Error 'Table 'tmp_replication_stop' already exists' on query. Default database: 'mercari'. Query: 'CREATE TABLE tmp_replication_stop ...
```

  ## Install

Please download release page or `mkr plugin install kazeburo/go-check-mysql-slave-sql-error`.
