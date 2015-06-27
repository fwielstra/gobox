
# maek poast. ipv4 for localhoast connections on OSX.
curl --ipv4 -H "Content-type:application/json" --data @json/testpoast.json http://localhost:8080/poast

# get poasts
curl --ipv4 http://localhost:8080/poast
