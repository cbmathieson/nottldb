echo "Stopping redis instances..."

folder=ports/
if [ -d "$folder" ]
    then
    input=ports/redis.txt
    if [ ! -f $input ]
        then
        echo "can't find instances to close"
    else
        echo "Removing redis containers..."
        instances="0"
        while IFS= read -r line
        do
            let "instances++"
            docker stop "redis$instances"
            docker rm "redis$instances"
        done < "$input"
        rm ports/redis.txt
    fi
else
    echo "can't find instances to close"
fi
