/*
Package awot provides an API for collecting public keys in a decentralized fashion on a network.

The name "awot" is short for Automatic Web of Trust.
As its name suggests, awot is based on a Web of Trust model for sharing and collecting public keys.
Contrary to the PGP Web of Trust, awot is automated : it does not require human interaction once loaded.
Since there is no required human validation when collecting keys, it is not completely safe from possible attacks.
However it tries to solve these problems by computing releveant confidence levels for each obtained key, this can help avoiding key collisions or impersonations.
Package awot is best used in addition to a reputation system in a network, a system that can ouput a "trust" level for each peer, that is how much trust we can put on this peer to share good public keys.
*/
package awot
