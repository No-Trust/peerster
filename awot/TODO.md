# TODO

Any new idea is welcomed.

## Main Features
- RSA PSS check the use of the salt
- Stop KeyRing

## Visualization
- Add color for edges function of the public key (if different public keys to the same peer, different colors)
- Implement an optional web-server displaying the keyring as done in Peerster's gossiper.
- Make the visualization dynamic.

## GUI
- Implement a GUI (webserver + frontend)  
  it may be a good idea to start from the visualization.

## Code quality
- key_ring.go needs re-factoring (to be more general) and tests.
- More test cases everywhere, favor coverage first depth later at this stage.
- Add bash tests to go tests for key_ring.go

## Named possible improvements
