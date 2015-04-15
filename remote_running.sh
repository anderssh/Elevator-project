#! /bin/bash
echo "Start copying"
scp  -r ~/Desktop/Elevator-project patrickf@129.241.187.143:~/Desktop
echo "Finished copying"
ssh patrickf@129.241.187.143
echo "Logged in"