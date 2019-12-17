rm -f ../ba/model/*.go
go build
ECHO "go build over"
./go-mysql-model-creator -conf=./ba.conf -dist=../ba/model -connect=ba_default -debug=true
go fmt ../ba/model/*.go