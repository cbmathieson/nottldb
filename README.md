# nottldb

A backend system for nottl

- Notes placed on the map are stored on presharded redis instances
- User information is stored on a Postgres database


## Redis Instances

Spinning them up
~~~
sh redis_db/run_redis.sh
~~~
Stopping them
~~~
sh redis_db/stop_redis.sh
~~~
