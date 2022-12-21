# GRPC CONTEXT CANCEL 

This repo contains a basic test harness to test the propagation and handling of context termination accross cascaded grpc calls, simulating multiple services connected over GRPC.

The tool can be used in **Client**, **Relay** or **Termination** mode. The typical setup is :

```mermaid
graph LR
    Client-->Relay1;
    Relay1-->Relay2;
    Relay2-->Termination;
```

Any number of instances can be cascaded in **Relay** mode.

The binary can be built from /cmd/txrx. 


## Usage

Run with --help for additional params.

```bash
❯ ./txrx --help
Usage of ./txrx:
  -attempts int
    	number of attempts (default 5)
  -id int
    	id of service instance (default 1)
  -opmode string
    	mode of operation of service instance [client,relay,termination] (default "relay")
  -reportInterval int
    	stats reporting interval (in seconds) (default 5)
  -rxport int
    	server port (default 9000)
  -rxtimeout int
    	simulated time (in ms) server takes to process a call
  -txport int
    	target server port (default 9000)
  -txtimeout int
    	time (in ms) before the grpc client cancels the request context
```


## Sample output

### This is a sample output for a basic client -> termination setup

#### Client Side

```bash
./txrx -id 1 -opmode client -txport 10000 -attempts 5 -txtimeout 5
INFO[0000] starting in client mode
INFO[0000] sending message 1 to next service             service=client
ERRO[0000] Error while sending: rpc error: code = Canceled desc = context canceled  service=client
INFO[0000] sending message 2 to next service             service=client
ERRO[0000] Error while sending: rpc error: code = Canceled desc = context canceled  service=client
INFO[0000] sending message 3 to next service             service=client
ERRO[0000] Error while sending: rpc error: code = Canceled desc = context canceled  service=client
INFO[0000] sending message 4 to next service             service=client
ERRO[0000] Error while sending: rpc error: code = Canceled desc = context canceled  service=client
INFO[0000] sending message 5 to next service             service=client
ERRO[0000] Error while sending: rpc error: code = Canceled desc = context canceled  service=client
INFO[0000] Attempted to send 5 messages                  service=client
```


#### Server(termination) Side

```bash
❯ ./txrx -id 2 -opmode termination -rxport 10000 -rxtimeout 10
INFO[0000] starting in termination mode
INFO[0000] Server listening on port 10000                serverID=2 service=server
INFO[0002] Starting handler                              messageID=1 serverID=2 service=server
INFO[0002] Context terminated: context canceled          messageID=1 serverID=2 service=server
INFO[0002] Starting handler                              messageID=2 serverID=2 service=server
INFO[0002] Context terminated: context canceled          messageID=2 serverID=2 service=server
INFO[0002] Starting handler                              messageID=3 serverID=2 service=server
INFO[0002] Context terminated: context canceled          messageID=3 serverID=2 service=server
INFO[0002] Starting handler                              messageID=4 serverID=2 service=server
INFO[0002] Context terminated: context canceled          messageID=4 serverID=2 service=server
INFO[0002] Starting handler                              messageID=5 serverID=2 service=server
INFO[0002] Context terminated: context canceled          messageID=5 serverID=2 service=server
INFO[0005] Total messages received: 5                    serverID=2 service=server
```