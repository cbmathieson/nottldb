function is_power_of_two () {
    declare -i n=$1
    (( n > 0 && (n & (n - 1)) == 0))
}

instances="0"

while true; do

read -p '# of redis instances (must be 2^n): ' instances

if is_power_of_two "$instances"; then
    break
else
    echo "Please enter a value == 2^n"
fi

done

echo "Spinning up $instances instances..."

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

for ((i=1; i<=instances; i++))
do
_="$(docker run --name "redis$i" -d -P redis)"
fullport="$(docker port redis$i)"
port=${fullport#*:}
echo "redis$i is listening on port: $port"
echo "$port" >> "$file"
done
