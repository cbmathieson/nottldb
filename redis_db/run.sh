# 0 is true and 1 is false!!!!!
is_power_of_two () {
    val="$1"
    if [ "$val" -eq 0 ]
    then
        return 1
    else
        while [ ! "$val" -eq 1 ]
        do
            if (($val % 4 != 0))
            then
                return 1
            fi
            val=$(($val / 4))
        done
        return 0
    fi
}

instances="0"

while true; do

read -p '# of redis instances (must be 4^n): ' instances

if is_power_of_two "$instances"; then
    break
else
    echo "Please enter a value == 4^n"
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
