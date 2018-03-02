# Automated Web of Trust (awot)

## What it is
A fully decentralized solution to the Public-Key Distribution Problem.

## Why it exists
Using crypto is great, but advertising public keys in IP network is prone to impersonation or man in the middle attacks.
AWOT provides uses an implicit underlying social network to detect and avoid problems such as collision of keys (two keys for a given user) or sibyl attacks.
This allows to obtain public keys from trusted peers and redistribute them, while recording how much "trust" one put on those keys.
> Please check the [write up](https://github.com/No-Trust/doc/blob/master/doc/write_up.pdf) to fully understand the benefits and drawbacks of this method.

## How it works
[write up](https://github.com/No-Trust/doc/blob/master/doc/write_up.pdf)

# How to use
This library is to be used on an existing decentralized network.
Check out our Peerster to see a working implementation.

And please check out the "go doc" :)

You may need to get used to these objects :
- KeyRecord : A key and its owner's name.
- TrustedKeyRecord : A KeyRecord with a confidence level attached to it.
- KeyExchangeMessage : A message that contains every information needed for sharing and receiving public key association.
  These are the messages that will need to be sent and received in the network. It contains : the public key, the owner's name of the key, the sender's name of the message and the signature of the key with the owner name, signed by the sender.
- KeyRing : this is the main database that will need to be updated with the received KeyExchangeMessages, it will perform some computations and gives back the trusted keys and confidence levels. It needs to be started, and will spawn a thread. 
