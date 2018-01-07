#!/usr/bin/env bash

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m'
DEBUG="false"

file3=file3.jpg
hash3=b4a4936e0f3153ea38b787eead84a79f5b17e806a4c09d0546903445607ee91b

startGossip(){
	local name=$1
	local port=$2
	local uiport=$3
	local keysf=$4
	local peers=""
	if [ "$1" != "A" ]; then
		# A -> B,C,D
		peers="-peers=127.0.0.1:5001,127.0.0.1:5002,127.0.0.1:5003"
	elif [ "$1" != "B" ]; then
		# B -> C,G
		peers="-peers=127.0.0.1:5002,127.0.0.1:5006"
	elif [ "$1" != "D" ]; then
		# D -> E,F
		peers="-peers=127.0.0.1:5004,127.0.0.1:5005"
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

startGossip A 5000 10000 ../../keysA/
startGossip B 5001 10001 ../../keysB/
startGossip C 5002 10002 ../../keysC/
startGossip D 5003 10003 ../../keysD/
startGossip E 5004 10004 ../../keysE/
startGossip F 5005 10005 ../../keysF/
startGossip G 5006 10006 ../../keysG/

sleep 5

echo -e "${GREEN}G is given file3${NC}"
./cli -UIPort=10006 -file=$file3

sleep 2
