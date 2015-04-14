#! /bin/bash
echo "Start copying"
scp  -r ~/Desktop/Elevator-project anderssh@129.241.187.154:~/Desktop
echo "Finished copying"
ssh anderssh@129.241.187.154
echo "Logged in"