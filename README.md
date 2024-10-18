This README provides instructions to set up our implementation of the Kademlia Distributed Hash Table (DHT) using Docker, along with steps to run tests and view the code coverage using Go.
Kademlia is a peer-to-peer (P2P) protocol for decentralized storage and retrieval of data, which enables efficient and scalable network communication. This implementation demonstrate the fundamental principles of Kademlia, including setting up a network, finding nodes and store and retrieve data. A command line interface (CLI) was also implemented to make it easier for users to communicate with the system.

## Prerequisites
* Docker
* Go
 
## Installation

1. Clone the repository:
```bash
git clone git@github.com:Limeiz/D7024E_Kademlia.git
cd D7024E_Kademlia
```
2. Build the Docker image
```bash
docker build -t kadlab:latest .
```
3. Start the Kademlia nodes
```bash
docker-compose up -d
```
4. Access a running node. In this example, we access kademlia-node-1
```bash
docker exec -it kademlia-node-1 /bin/sh
```

## Test the implementation
1. Run the Tests
```bash
go test -coverprofile cover.out
```
2. View the code coverage in your browser
```bash
go tool cover -html=cover.out
```

## Stop running the containers
```bash
docker-compose down --remove-orphans 
```

## CLI
Once inside a node, you can interact with the Kademlia network using the following commands, all prefixed with kademlia
```bash
ping <to>
put <data>
get <hash>
show-id
show-storage
show-routing-table
forget <hash>
exit
help
```
