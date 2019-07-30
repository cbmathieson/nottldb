# 0 is true and 1 is false!!!!!
function is_power_of_two () {
    declare -i n=$1
    if ((n == 0))
    then
        return 1
    else
        while ((n != 1))
        do
            if ((n % 4 != 0))
            then
                return 1
            fi
            n=$((n / 4))
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
