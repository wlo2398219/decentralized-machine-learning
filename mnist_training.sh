#!/usr/bin/env bash

cd ../finalproject
go build 
cd client 
go build 
cd ../../demo
cp ../finalproject/finalproject .
cp ../finalproject/client/client .

name='A'

for i in `seq 1 10`;
do
	mkdir -p $name
	mkdir -p "$name/_Datasets"
	cp finalproject "$name"
	cp client "$name"

	# if [[ $i > 1 ]]; then
		#statements
	cp "hidden_layer_train_split/hidden_layer_train_$((i-1)).csv" "$name/_Datasets/hidden_layer_train.csv"
	# fi

	name=$(echo "$name" | tr "A-Y" "B-Z")

done

UIPort=10000
gossipPort=5000
name='A'

for i in `seq 1 10`;
do
	cd $name
	outFileName="$name.out"
	peerPort=$((($gossipPort+1)%10+5000))
	peer="127.0.0.1:$peerPort"
	gossipAddr="127.0.0.1:$gossipPort"
	./finalproject -UIPort=$UIPort -gossipAddr=$gossipAddr -name=$name -peers=$peer -rtimer=1 > $outFileName &
	outputFiles+=("$outFileName")

	echo "$name running at UIPort $UIPort and gossipPort $gossipPort"
	UIPort=$(($UIPort+1))
	gossipPort=$(($gossipPort+1))
	name=$(echo "$name" | tr "A-Y" "B-Z")

	cd ..
done

sleep 10

./A/client -UIPort=10000 -train -file="mnist"




# pkill -f finalproject