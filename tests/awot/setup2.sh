#!/usr/bin/env bash

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m'
DEBUG="false"

mkdir -p keys

cd ../../gossiper
go build
cd ../cli
go build
cd ../tests/awot

mkdir -p keys0 keys1 keys2 keys3 keys4 keys5
mkdir -p peer0 peer1 peer2 peer3 peer4 peer5
mkdir -p peer0/gossiper peer1/gossiper peer2/gossiper peer3/gossiper peer4/gossiper peer5/gossiper
cp ../../gossiper/gossiper peer0/gossiper/gossiper
cp ../../gossiper/gossiper peer1/gossiper/gossiper
cp ../../gossiper/gossiper peer2/gossiper/gossiper
cp ../../gossiper/gossiper peer3/gossiper/gossiper
cp ../../gossiper/gossiper peer4/gossiper/gossiper
cp ../../gossiper/gossiper peer5/gossiper/gossiper
cp ../../cli/cli ./

startGossip(){
	local name=$1
	local port=$2
	local uiport=$3
	local keysf=$4
	local peers=""
	if [ "$1" != "peer1" ]; then
		peers="-peers=127.0.0.1:5001"
	elif [ "$1" != "peer3" ]; then
		peers="-peers=127.0.0.1:5004"
	elif [ "$1" != "peer4" ]; then
		peers="-peers=127.0.0.1:5005"
	fi

	echo ./gossiper -gossipAddr=127.0.0.1:$port -UIPort=$uiport -name=$name $peers

	cd $1
	# delete previous downloads
	rm -rf _Downloads/
	cd gossiper
	# delete log
	rm -f *.log
	# launch gossiper
	./gossiper -gossipAddr=127.0.0.1:$port -UIPort=$uiport -name=$name $peers> $name.log -keys=$keysf &
	cd ../..

	# don't show 'killed by signal'-messages
	disown
}

startGossip peer0 5000 10000 ../../keys0/
startGossip peer1 5001 10001 ../../keys1/
startGossip peer2 5002 10002 ../../keys2/
startGossip peer3 5003 10003 ../../keys3/
startGossip peer4 5004 10004 ../../keys4/
startGossip peer5 5005 10005 ../../keys5/

sleep 31
killall gossiper

cp peer0/peer0.pub keys/
cp peer1/peer1.pub keys/
cp peer2/peer2.pub keys/
cp peer3/peer3.pub keys/
cp peer4/peer4.pub keys/
cp peer5/peer5.pub keys/

cp peer1/peer1.pub keys0/

cp peer0/peer0.pub keys1/
cp peer2/peer2.pub keys1/
cp peer3/peer3.pub keys1/

cp peer1/peer1.pub keys2/
cp peer4/peer4.pub keys2/

cp peer1/peer1.pub keys3/
cp peer4/peer4.pub keys3/

cp peer5/peer5.pub keys4/

#testing
fail(){
	echo -e "${RED}*** Failed test $1 ***${NC}"
  exit 1
}
