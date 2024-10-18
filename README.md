# New readme
This README provides instructions to set up a Kademlia Distributed Hash Table (DHT) implementation using Docker, along with steps to run tests and view the code coverage in Go.
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

# Old readme
## Some tips

Here's some advice I like to give students during sprint 0.
But it's probably better to put it in one place so everyone gets the same information and so that you can refer back to it later.


### Additional resources

The PDF describing the lab assignment (in Canvas) contains a link to a web page in a footnote.
This page contains helpful information on implementation details that are missing in the original paper that describes Kademlia.
Because it is easy to miss, I'm also including it here:
[https://xlattice.sourceforge.net/components/protocol/kademlia/specs.html](https://xlattice.sourceforge.net/components/protocol/kademlia/specs.html)


### Testing

Achieving a minimum test coverage of 50 % is specified as a separate mandatory requirement (M4).
However, I strongly suggest that you don't treat it as a separate task to work on.
That is, **don't implement first and then add tests at the end!**

First of all, achieving the mandatory coverage will be much more difficult if you add tests after finishing the implementation.
Second, your tests should help you get the implementation working in the first place!
So look at the tests as a tool to help you rather than a box to tick.
(In fact, if you write tests in parallel with the implementation, then the optional (non-mandatory) goal of 80 % (U5) should not be very difficult to achieve.)


### Concurrency and thread safety

Make sure you really understand race conditions, so that you know what problem it is you are trying to solve.
Think carefully about which parts of the code are run by more than one thread (or goroutine) and, in particular, which variables/datastructures are accessed by different threads.

You can use old-fashioned locks to control access to critical regions.
That's a perfectly valid solution and perhaps one you are familiar with and find natural to think about.
However, there are other options, especially in Go, which has *channels* built into the language.
For example, one option is to let only a single goroutine have access to a particular datastructure, and then other goroutines communicate with it using channels.
If you get used to this way of thinking, you may find this solution to be **simpler** than locks!

Either way, you can run your tests with the `-race` flag (see documentation: [Data Race Detector](https://go.dev/doc/articles/race_detector)).
This will instrument your code so that race conditions can be detected automatically.
**However**, how helpful this is depends on how good your tests are.
If there are race conditions that your tests never touch, then they will not be detected.
(Running with `-race` is a dynamic rather than static analysis!)


### Report

The instructions say that the report needs to include "a system architecture description that also contains an implementation overview".
A common question is what this should look like.
Ultimately, the point is that you are supposed to communicate to someone how your system is designed.
What are the main components, and how do they communicate with each other, etc?
Just like you have to make choices about what the best way is to design and implement your solution, you have to make choices about what the best way is to communicate to someone else what you have done.

Try this: imagine that we change the assignment so that the students next year will get your implementation as a starting point and are asked to improve it, such as adding features.
*What information would they need so they quickly understand your implementation and can start modifying it?*
*What design choices should they be aware of?*
