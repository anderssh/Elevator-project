#! /bin/bash
export GOPATH=$(pwd)
echo "Building executable"
go build main.go
echo "Start copying"
scp  -r ~/Desktop/Elevator-project/main patrickf@129.241.187.146:~/Desktop
echo "Finished copying"
ssh patrickf@129.241.187.146
echo "Logged in"