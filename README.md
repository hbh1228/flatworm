# FlatWorm Repository

Asynchronous fault-tolerant protocols for the following paper: 

Baohan Huang, Haibin Zhang, Sisi Duan, Boxin Zhao, and Liehuang Zhu. "Practical Asynchronous BFT from Local Coins"

This repository implements one BFT protocols:

+ FlatWorm


Different RBC modules are implemented under src/broadcast, and different ABA modules are implemented under src/aba. See the README files under each folder for details.

### Configuration

The codebase can only be compiled for x86 machines for now. The src/cryptolib/word/word_amd64.s needs to be changed to adapt to arm machines. 

Configuration is under etc/conf.json

Change "consensus" to switch between the protocols. See note.txt for details. 

### Experimental environment && Dependency

+ system: ubuntu 20.04
+ golang: go1.20 linux/amd64
+ github.com/cbergoon/merkletree v0.2.0
+ github.com/golang/protobuf v1.5.4
+ github.com/klauspost/reedsolomon v1.12.3
+ google.golang.org/grpc v1.65.0
+ google.golang.org/protobuf v1.34.2
+ github.com/klauspost/cpuid/v2 v2.2.6
+ golang.org/x/net v0.25.0
+ golang.org/x/sys v0.20.0
+ golang.org/x/text v0.15.0
+ google.golang.org/genproto/googleapis/rpc v0.0.0-20240528184218-531527333157

#### How to run the code 

+ enter the directory and run the following commands
```
export GOPATH=$PWD
export GOBIN=$PWD/bin
export GO111MODULE=on
```

+ If you only need to build again, run 
```
go build src/main/server.go
go build src/main/client.go
go build src/main/keygen.go
```

##### Launch the code

+ Modify the configuration file  "etc/conf.json" to choose which protocol to execute, and modify the IP addresses 
and port numbers of all servers. Details about the protocols are included in "$etc/node.txt$". The “id” of each server should be unique. 
By default, we use monotonically increasing ids, 0, 1, 2, ....


+ For all the servers, run the command below to start the servers:
```
         server [id]
```
Here, "id" is configured in conf.json and is different at each server.
+ Start a client to send transaction to start the protocol by running the following command:
```
        client [id] 1 [batch] [msg]
```
Here, "id" is the identifier of the client. We do not require the client to be registered. One can use any id that is unique,
e.g., 1000. "b" is the batch size. "msg" can be any message. One can ignore the "msg" field and a default message is included in the codebase.

##### Quick experiment locally
+ Open four terminals and start server in each terminal (start one server in one terminal).
```
        server 0
        server 1
        server 2
        server 3
```
+ Open one terminal to start client and send transactions.
```
        client 100 1 10000
```
All server terminal will print text like " *****epoch ends", which represents the success of the epoch.
One can repeat the operation of client after the epoch end.
