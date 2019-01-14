#!/usr/bin/env bash

mnistDir="../finalproject/_Datasets/mnist"

cd ../finalproject
go build 
cd client 
go build 
cd ../../demo
cp ../finalproject/finalproject .
cp ../finalproject/client/client .
cp ../finalproject/rand1.py .

# Initialize feature extractor.
# Add ../ before $mnistDir because the generated file will be moved to ./$name
echo "Initialize feature extractor."
bash $mnistDir/mnist_feature_init.sh "../$mnistDir"

name="A"
mode="$1"
newpeers="$2"
byzantineMode="$3"

for i in `seq 1 10`;
do
	mkdir -p $name
	mkdir -p "$name/_Datasets"
	cp finalproject "$name"
	cp client "$name"
	cp mnist_feature_extractor.sh "$name"

	if [[ $i == 1 ]]; then
		#statements
		cp "hidden_layer_test.csv" "./A/_Datasets/"
		cp "../finalproject/index.html" "./A/"

	fi
	cp "hidden_layer_train_split/hidden_layer_train_$((i-1)).csv" "$name/_Datasets/hidden_layer_train.csv"



	name=$(echo "$name" | tr "A-Y" "B-Z")

done

UIPort=10000
gossipPort=0
name='A'
nNode=10


if [[ $newpeers == "Y" ]]; then
	python3 rand1.py "DABCD" "$mode" "" "7"	
else 
	if [[ $byzantineMode == "Y" ]]; then
		python3 rand1.py "DABCD" "$mode" "E" "10"
	else
		python3 rand1.py "DABCD" "$mode" "" "10"
	fi
fi
# for i in `seq 1 $nNode`;
# do
# 	cd $name
# 	outFileName="$name.out"
# 	peerPort=$((($gossipPort+1)%$nNode+5000))
# 	peer="127.0.0.1:$peerPort"
# 	gossipAddr="127.0.0.1:$(($gossipPort+5000))"
# 	echo "$name running at UIPort $UIPort and gossipAddr $gossipAddr and peer $peer"
# 	python3 rand.py "$name" "1"
# 	# ./finalproject -UIPort=$UIPort -gossipAddr=$gossipAddr -name=$name -peers=$peer -rtimer=5 > $outFileName &
# 	outputFiles+=("$outFileName")

# 	UIPort=$(($UIPort+1))
# 	gossipPort=$(($gossipPort+1))
# 	name=$(echo "$name" | tr "A-Y" "B-Z")

# 	cd ..
# done

sleep 10

# ./A/client -UIPort=10000 -train -file="mnist"
./A/client -UIPort=10000 -train -file="mnist"

sleep 10

if [[ $newpeers == "Y" ]]; then
	cd J
	./finalproject -UIPort=10009 -gossipAddr=5010 -name=J -peers=127.0.0.1:5002 -rtimer=5 -mode=distributed -byz=False > J.out &
	cd ..

	sleep 5

	cd H
	./finalproject -UIPort=10007 -gossipAddr=127.0.0.1:5007 -name=H -peers=127.0.0.1:5000 -rtimer=5 -mode=distributed -byz=False > H.out &
	cd ..

	sleep 5

	cd I
	./finalproject -UIPort=10008 -gossipAddr=127.0.0.1:5008 -name=I -peers=127.0.0.1:5002 -rtimer=5 -mode=distributed -byz=False > I.out &
	cd ..
fi


read -p "Press [Enter] key to test.."
for i in `seq 1 10`;
do
	read num
	if [[ $num == "-1" ]]; then
		#statements
		break
	fi
	./A/client -UIPort=10000 -test -file="images/$num.png"
	# ./A/client -UIPort=10000 -test -file="images/0.png"
done

read -p "Press [Enter] key to stop.."
pkill -f finalproject
