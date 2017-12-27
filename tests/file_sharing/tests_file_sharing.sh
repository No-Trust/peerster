#!/usr/bin/env bash

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m'
DEBUG="false"

cd ../../gossiper
go build
cd ../cli
go build
cd ../tests/file_sharing

mkdir -p peer0 peer1 peer2
mkdir -p peer0/gossiper peer1/gossiper peer2/gossiper
cp ../../gossiper/gossiper peer0/gossiper/gossiper
cp ../../gossiper/gossiper peer1/gossiper/gossiper
cp ../../gossiper/gossiper peer2/gossiper/gossiper
cp file1.txt peer0/gossiper/
cp file3.jpg peer0/gossiper/
cp file2.txt peer1/gossiper/
cp ../../cli/cli ./

outputFiles=()
message1=Winter_Is_Here_!
message2=iN1Xha3OK09UvtLYVJghL8PaVVu

file1=file1.txt
hash1=f43aef47d6b4ad22a6cca344a4d58fd3775eec47e1dc57b4e2752c98bf2daf2f

file2=file2.txt
hash2=5e12077a3d13bfc677cb0feb5aa9ebe6728b7d2f404b0f4426eb1fc5a4f06d16

file3=file3.jpg
hash3=b4a4936e0f3153ea38b787eead84a79f5b17e806a4c09d0546903445607ee91b

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
	./gossiper -gossipAddr=127.0.0.1:$port -UIPort=$uiport -name=$name $peers> $name.log &
	cd ../..

	# don't show 'killed by signal'-messages
	disown
}

startGossip peer0 5000 10000
startGossip peer1 5001 10001
startGossip peer2 5002 10002

sleep 14

# submit file1.txt and file3.jpg to peer0
echo -e "${GREEN}peer0 is given file1 and file3${NC}"
./cli -UIPort=10000 -file=$file1
./cli -UIPort=10000 -file=$file3

# submit file2.txt to peer1
echo -e "${GREEN}peer1 is given file2${NC}"
./cli -UIPort=10001 -file=$file2

sleep 2

# request file1.txt from peer0 at peer1
echo -e "${GREEN}peer1 asks peer0 for file1${NC}"
./cli -UIPort=10001 -file=myfile1.txt -request=$hash1 -Dest=peer0

# request file2.txt from peer1 at peer0
echo -e "${GREEN}peer0 asks peer1 for file2${NC}"
./cli -UIPort=10000 -file=myfile2.txt -request=$hash2 -Dest=peer1

sleep 2

# request file1.txt from peer2 at peer1
echo -e "${GREEN}peer2 asks peer1 for file1${NC}"
./cli -UIPort=10002 -file=myfile1.txt -request=$hash1 -Dest=peer1

sleep 5

killall gossiper

#testing
fail(){
	echo -e "${RED}*** Failed test $1 ***${NC}"
  exit 1
}


# check that file1.txt is downloaded at peer1 and peer2
grep -q "$message1" peer1/_Downloads/myfile1.txt || fail "peer1 did not get file1.txt from peer0"
grep -q "$message1" peer2/_Downloads/myfile1.txt || fail "peer2 did not get file1.txt from peer1"

# check that file2.txt is downloaded at peer0
grep -q "$message2" peer0/_Downloads/myfile2.txt || fail "peer0 did not get file2.txt from peer2"

rm -rf peer0 peer1 peer2 cli
