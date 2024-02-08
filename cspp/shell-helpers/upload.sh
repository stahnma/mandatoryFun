#curl -X POST -F "image=@/path/to/your/image.jpg" -F "caption=OptionalCaption" http://localhost:8080/upload

#file=`ls ~/Desktop/*.png ~/Desktop/*.jpeg |sort -R |tail -1`
if [ -z "$1" ]; then
	echo "Usage: $0 <image file> <caption>"
	exit 1
fi
if [ -z "$2" ]; then
	echo "Usage: $0 <image file> <caption>"
	exit 1
fi
file="$1"
caption="$2"

curl -X POST \
  -F "image=@$file" \
  -F "caption=$caption"  \
  -H "X-API-KEY: $API_KEY" \
http://localhost:7171/upload


