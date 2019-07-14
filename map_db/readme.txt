The map database is using multiple (fixed) redis instances as data stores. Since redis as a data store does not play well with sharding after launch, I'm going to preshard 128 instances to cover the surface of the earth.

These instances will be spread evenly between -85 to 85 lat / -180 to 180 long. Maybe in the future I'll spread them out based on phone usage density. There will be a 10km overlap between each instance, so if a user falls between both of them it gets uploaded to both and once found, gets a set expiry on both of them

A redis instance that does not hold any information is very small (1 MB). So for my computers specs (8 GB RAM) I'm going to create enough instances to push the system to 4GB usage with varying amounts of load:

1 instance at 4GB
2 instances at 2GB,
4 instances at 1GB,
16 instances at 250MB,
32 instances at 125MB,
64 instances at 62.5MB,
128 instances at 31.25MB

* note: the capacity includes instances not in use *

page 147 gives an example of startning

