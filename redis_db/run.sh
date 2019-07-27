echo "Spinning up redis instances..."

file=redis_ports.txt
if [ -f "$file" ]
then
    rm redis_ports.txt
fi

touch redis_ports.txt

for i in {1..4}
do
_="$(docker run --name "redis$i" -d -P redis)"
fullport="$(docker port redis$i)"
port=${fullport#*:}
echo "redis$i is listening on port: $port"
echo "$port" >> "$file"
done
