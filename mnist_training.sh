#!/usr/bin/env bash

mnistDir="../finalproject/_Datasets/mnist"

cd ../finalproject
go build 
cd client 
go build 
cd ../../demo
cp ../finalproject/finalproject .
cp ../finalproject/client/client .

# Initialize feature extractor.
# Add ../ before $mnistDir because the generated file will be moved to ./$name
echo "Initialize feature extractor."
bash $mnistDir/mnist_feature_init.sh "../$mnistDir"

name='A'

for i in `seq 1 10`;
do
	mkdir -p $name
	mkdir -p "$name/_Datasets"
	cp finalproject "$name"
	cp client "$name"
	cp mnist_feature_extractor.sh "$name"

	# if [[ $i > 1 ]]; then
		#statements
	cp "hidden_layer_train_split/hidden_layer_train_$((i-1)).csv" "$name/_Datasets/hidden_layer_train.csv"
	# fi

	name=$(echo "$name" | tr "A-Y" "B-Z")

done

UIPort=10000
gossipPort=0
name='A'
nNode=10

for i in `seq 1 $nNode`;
do
	cd $name
	outFileName="$name.out"
	peerPort=$((($gossipPort+1)%$nNode+5000))
	peer="127.0.0.1:$peerPort"
	gossipAddr="127.0.0.1:$(($gossipPort+5000))"
	echo "$name running at UIPort $UIPort and gossipAddr $gossipAddr and peer $peer"
	./finalproject -UIPort=$UIPort -gossipAddr=$gossipAddr -name=$name -peers=$peer -rtimer=5 > $outFileName &
	outputFiles+=("$outFileName")

	UIPort=$(($UIPort+1))
	gossipPort=$(($gossipPort+1))
	name=$(echo "$name" | tr "A-Y" "B-Z")

	cd ..
done

sleep 10

# ./A/client -UIPort=10000 -train -file="mnist"
./A/client -UIPort=10000 -train -file="mnist"

read -p "Press [Enter] key to test.."
./A/client -UIPort=10000 -test -file="images/0.png"

read -p "Press [Enter] key to stop.."
pkill -f finalproject
