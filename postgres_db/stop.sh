echo "Stopping user_db..."

_="$(docker stop user_db)"
_="$(docker rm user_db)"

rm ports/db.txt

echo "user_db is stopped"
