if [[ -z $2 ]]; then
  echo "specify client connection and count"
else
  for ((i=1;i<=$2;i++)); do
    go run main.go c $1 &
  done
fi