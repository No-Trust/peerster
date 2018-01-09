# Peerster

## Important notes
NAT traversal, File search and Name change have been deactivated for our project. Our project does not include these functionalities.

## Usage

firstly,
> go build

for gossiper/ gui/ 

### Gossiper

in /peerster/gossiper, example of (local) use :
> ./gossiper -UIPort=10000 -gossipAddr=127.0.0.1:5000 -keys=../../keysA -name=A -peers=127.0.0.1:5001

The only change from previous homeworks specifications for the options are the keys flag.
This flag is important and must be given. To launch a peer named A, we recommand creating a folder named keysA and place all fully trusted keys in it, with filenames like 'peerName.pub' where 'peerName' is the name of the peer in the network having this public key.
The filename is necessary for the system to work as intended as it gives the name of the peer associated with the key.
Then please link the folder to the gossiper by specifying its path in the flag -keys=_ .

When launching a gossiper, in the upper folder it will check for the presence of files private.key and peerName.pub, if it does not see these files, the gossiper will generate a new key and save the files.

Therefore, we recommand to launch a first time the gossipers for them to generate the keys, and then to relauch them and share the keys.

#### Testing

For testing, please check our test files setup.sh, launch.sh and clean.sh in folder tests/authentic_file_sharing.
setup.sh is used to generate the keys. It may fail to finish properly if the processor is not fast enough. In this case please run clean.sh and change the sleep value in  setup.sh to something bigger. (Generating keys may take a long time)

##### Gossipers

> sh setup.sh

Will generate the folder structure and the keys.

> sh launch.sh

Will instantiate the gossipers and give to each of them some keys of the others to sign (bootstrap).

> sh clean.sh

Will kill the gossiper processes and delete the created folders and files in setup.sh.


##### Gui
in /peerster/gui :
> ./gui -UIPort=10000 -port=8080

Then request localhost:8080 in your browser.
The keyring visualization can be viewed at localhost:8080/keyring


The gui allows the user to send private messages when clicking on one peer name in the chats tab. The private messages sent are not showed to the sender.

One can input a file that has to be present at the folder where the gossiper has been lauched.

One can also request a file to be downloaded and has to specify the name of the file (local name), the identifier of the file (SHA256), the origin of the file (the name of the peer that put this file on the network) and host that will be sent the request (the uploader of the file).
Any of these fields is necessary.

If the file is present on Host, and effectively created by Origin, the file will be downloaded in \_Downloads as before. If the download is not authenticated (we cannot certify the file comes from Host or is not signed by Origin), the download will not succeed.

Please note that no notifications are sent to the gui yet, we apologize for this. However any notification is present in the log files or standard output.

##### CLI
Same as for previous homeworks.

For more details :
> ./gossiper -h
> ./gui -h
> ./cli -h
