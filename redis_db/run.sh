echo "Spinning up redis instances..."

folder=ports/

if [ ! -d "$folder" ]
then
mkdir "$folder"
fi

file=ports/redis.txt
if [ -f "$file" ]
then
    rm "$file"
fi

touch "$file"

for i in {1..4}
do
_="$(docker run --name "redis$i" -d -P redis)"
fullport="$(docker port redis$i)"
port=${fullport#*:}
echo "redis$i is listening on port: $port"
echo "$port" >> "$file"
done
