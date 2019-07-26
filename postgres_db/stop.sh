echo "Stopping user_db..."

_="$(docker stop user_db)"
_="$(docker rm user_db)"

echo "user_db is stopped"
