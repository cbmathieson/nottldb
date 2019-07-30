echo "Stopping redis instances..."

folder=ports/
if [ -d "$folder" ]
    then
    input=ports/redis.txt
    if [ ! -f $input ]
        then
        echo "can't find instances to close reeeee"
    else
        echo "Removing redis containers..."
        instances="0"
        while IFS= read -r line
        do
            let "instances++"
        done < "$input"
        for ((i=1; i<=instances; i++))
        do
            docker stop "redis$i"
            docker rm "redis$i"
        done
        rm ports/redis.txt
    fi
else
    echo "can't find instances to close yeeee"
fi
