#!/usr/bin/env bash

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m'
DEBUG="false"

# compiling
cd ../../gossiper
go build
cd ../cli
go build
cd ../gui
go build
cd ../tests/authentic_file_sharing/

# creating folders for different peers
mkdir -p keys
mkdir -p keysA keysB keysC keysD keysE keysF keysG
mkdir -p A B C D E F G
mkdir -p A/gossiper B/gossiper C/gossiper D/gossiper E/gossiper F/gossiper G/gossiper
# copying binaries
cp ../../gossiper/gossiper A/gossiper/gossiper
cp ../../gossiper/gossiper B/gossiper/gossiper
cp ../../gossiper/gossiper C/gossiper/gossiper
cp ../../gossiper/gossiper D/gossiper/gossiper
cp ../../gossiper/gossiper E/gossiper/gossiper
cp ../../gossiper/gossiper F/gossiper/gossiper
cp ../../gossiper/gossiper G/gossiper/gossiper
cp ../../cli/cli ./
cp ../../gui/gui ./
mkdir -p public
cp ../../gui/public/* ./public/

startGossip(){
	local name=$1
	local port=$2
	local uiport=$3
	local keysf=$4
	local peers=""
	if [ "$1" == "A" ]; then
		# A -> B,C,D
		peers="-peers=127.0.0.1:5001,127.0.0.1:5002,127.0.0.1:5003"
	elif [ "$1" == "B" ]; then
		# B -> C,G
		peers="-peers=127.0.0.1:5002,127.0.0.1:5006"
	elif [ "$1" == "D" ]; then
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

sleep 40
pkill -f gossiper

cp A/A.pub keys/
cp B/B.pub keys/
cp C/C.pub keys/
cp D/D.pub keys/
cp E/E.pub keys/
cp F/F.pub keys/
cp G/G.pub keys/

# A <- keys B,C
cp keys/B.pub keysA/
cp keys/C.pub keysA/

# B <- keys D,E
cp keys/D.pub keysB/
cp keys/E.pub keysB/

# C <- keys E,F
cp keys/E.pub keysC/
cp keys/F.pub keysC/

# D <- keys G
cp keys/G.pub keysD/

# E <- keys G, A
cp keys/G.pub keysE/
cp keys/A.pub keysE/

# F <- keys G
cp keys/G.pub keysF/

# G <- keys E
cp keys/E.pub keysG/

# G is given file3.jpg
# b4a4936e0f3153ea38b787eead84a79f5b17e806a4c09d0546903445607ee91b
cp file3.jpg G/gossiper/
# G is given file2.txt
# 5e12077a3d13bfc677cb0feb5aa9ebe6728b7d2f404b0f4426eb1fc5a4f06d16
cp file2.txt G/gossiper/
