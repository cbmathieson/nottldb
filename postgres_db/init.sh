echo "Starting user database..."

docker run -d --name user_db -e POSTGRES_PASSWORD=pass -v my_dbdata:/var/lib/postgresql/data -P postgres:11

fullport="$(docker port user_db)"
port=${fullport#*:}
echo "user_db is listening on port: $port"
echo "password: pass"

echo "initialising user_db..."
_="$(psql -h localhost -U postgres -p $port -d postgres -f ./postgres_db/init_user_db.sql)"


