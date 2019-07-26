echo "Stopping redis instances..."

for i in {1..4}
do
docker stop "redis$i"
done

echo "Removing redis containers..."

for i in {1..4}
do
docker rm "redis$i"
done
