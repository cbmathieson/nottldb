# nottldb

A backend system for nottl

- Notes placed on the map are stored on presharded redis instances
- User information is stored on a Postgres database
- Go web server mediates requests


## Redis Instances

Spinning them up
~~~
sh redis_db/run.sh
~~~
Stopping them
~~~
sh redis_db/stop.sh
~~~

## Postgres Database

Spinning it up
~~~
sh postgres_db/init.sh
~~~
Stopping it
~~~
sh postgres_db/stop.sh
~~~

## Go Server

Spinning it up
~~~
sh go_server/run.sh
~~~
