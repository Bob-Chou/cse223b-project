#!/bin/bash

if [ _$GOPATH = _ ]; then
  echo "set GOPATH to root of project first"
  exit
fi

cd $GOPATH/src/visual

# install grpc
echo "preparing gRPC package..."
go get -u github.com/golang/protobuf/protoc-gen-go
go get -u github.com/golang/protobuf/proto
go get -u google.golang.org/grpc

# build go proto
echo "build protobuf for go..."
[ ! -d $GOPATH/src/go_protoc ] && mkdir $GOPATH/src/go_protoc
protoc --go_out=plugins=grpc:$GOPATH/src/go_protoc/ chord_message.proto

# build python proto
echo "preparing python package..."
pip install -r requirements.txt
echo "build protobuf for python..."
python -m grpc_tools.protoc --python_out=. --grpc_python_out=. -I. chord_message.proto

echo "done!"
cd -