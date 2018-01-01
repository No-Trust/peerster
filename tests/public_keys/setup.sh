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
cd ../tests/public_keys

mkdir -p peer0 peer1 peer2
mkdir -p peer0/gossiper peer1/gossiper peer2/gossiper
cp ../../gossiper/gossiper peer0/gossiper/gossiper
cp ../../gossiper/gossiper peer1/gossiper/gossiper
cp ../../gossiper/gossiper peer2/gossiper/gossiper
cp ../../cli/cli ./

# Making a simple network:
# 0 -> 1 -> 2
# 0 {file1.txt, file3.jpg}
# 1 {file2.txt}
# 2 {}

startGossip(){
	local name=$1
	local port=$2
	local uiport=$3
	local peers=""
	if [ "$1" == "peer0" ]; then
		peers="-peers=127.0.0.1:5001"
	else
		peers="-peers=127.0.0.1:5002"
	fi

	echo ./gossiper -gossipAddr=127.0.0.1:$port -UIPort=$uiport -name=$name $peers

	cd $1
	# delete previous downloads
	rm -rf _Downloads/
	cd gossiper
	# delete log
	rm -f *.log
	# launch gossiper
	./gossiper -gossipAddr=127.0.0.1:$port -UIPort=$uiport -name=$name $peers> $name.log -keys="../../keys/"&
	cd ../..

	# don't show 'killed by signal'-messages
	disown
}

startGossip peer0 5000 10000
startGossip peer1 5001 10001
startGossip peer2 5002 10002

sleep 10
killall gossiper

cp peer0/peer0.pub keys/
cp peer1/peer1.pub keys/
cp peer2/peer2.pub keys/

#testing
fail(){
	echo -e "${RED}*** Failed test $1 ***${NC}"
  exit 1
}
