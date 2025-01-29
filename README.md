# q2p

### node 1
````go run main.go -port 9000 -db ./data/node1````
### node 2
````go run main.go -port 9001 -db ./data/node2 -peer /ip4/127.0.0.1/tcp/9000/p2p````