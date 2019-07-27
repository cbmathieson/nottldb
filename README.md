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

Spinning it up (must be done after redis and db have started)
~~~
sh go_server/run.sh
~~~
All ports for redis (redis_ports.txt)
and the database port (db_port.txt) 
are printed to a text file to be read by the go server

