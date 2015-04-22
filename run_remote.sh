#! /bin/bash
export GOPATH=$(pwd)
echo "Building executable"
go build main.go
echo "Start copying"
scp  -r ~/Desktop/Elevator-project/main anderssh@129.241.187.150:~/Desktop
echo "Finished copying"
ssh anderssh@129.241.187.150
echo "Logged in"