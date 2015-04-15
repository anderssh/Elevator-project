#! /bin/bash
echo "Start copying"
scp  -r ~/Desktop/Elevator-project/main patrickf@129.241.187.141:~/Desktop
echo "Finished copying"
ssh patrickf@129.241.187.141
echo "Logged in"