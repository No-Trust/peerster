# TODO

Any new idea is welcomed.

## Main Features
- implement a threshold T for key confidence levels  
  If confidence < T, awot should report to the user no key, as if the key was absent, while keeping it stored in the key ring.  
  For now, it is the user that have to decide to use the key or not according to its confidence values.


## Visualization
- Implement an optional web-server displaying the keyring as done in Peerster's gossiper.
- Make the visualization dynamic.

## GUI
- Implement a GUI  
  it may be a good idea to start from the visualization.

## Code quality
- key_ring.go needs re-factoring (to be more general) and tests.
- More test cases everywhere.


## Named possible improvements
- key_record.go TrustedKeyRecord : use struct embedding over composition for the KeyRecord.
- key_ring.go KeyRing : use struct embedding over composition for the KeyTable.
- key_ring.go : split KeyRing implementation and visualization.
