go build
cd client
go build
cd ..

./finalproject -gossipAddr=127.0.0.1:5000 -UIPort=10000 -name=A -peers=127.0.0.1:5001 -rtimer=1 > A.out &
./finalproject -gossipAddr=127.0.0.1:5001 -UIPort=10001 -name=B -peers=127.0.0.1:5002 -rtimer=1 > B.out &
./finalproject -gossipAddr=127.0.0.1:5002 -UIPort=10002 -name=C -peers=127.0.0.1:5003 -rtimer=1 > C.out &
./finalproject -gossipAddr=127.0.0.1:5003 -UIPort=10003 -name=D -peers=127.0.0.1:5004 -rtimer=1 > D.out &
./finalproject -gossipAddr=127.0.0.1:5004 -UIPort=10004 -name=E -peers=127.0.0.1:5005 -rtimer=1 > E.out &
./finalproject -gossipAddr=127.0.0.1:5005 -UIPort=10005 -name=F -peers=127.0.0.1:5000 -rtimer=1 > F.out &
sleep 5

./client/client -UIPort=10000 -train -file="mnist"
# ./client/client -UIPort=10000 -train -file="uci_cbm_dataset.txt"


sleep 20

pkill -f finalproject

