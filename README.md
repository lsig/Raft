# Distributed Consensus - Raft

Welcome to the Raft implementation of group 5 in the course T-419-CADP at Reykjavik University. 
Below are instruction on how to run the code, what commands are avaible and more. 

## Server

```bash
# from ./
cd server
go run . <server-host>:<server-port> <filename>
```

## Server Commands

log: The log commands outputs the logs of the server to the stdout

print: The print commands displays useful information about the state of the server to the stdout

resume: Resume a suspended process

suspend: Suspend a process, moves it to a failed state

timeout: Artificially timesout the process, moves it to a candidate state


## Client

```bash
# from ./
cd client
go run . <server-host>:<server-port>
```

## Client Commands

exit: exits the program

## Acknowledgements

[Converting bytes to string](https://golangdocs.com/golang-byte-array-to-string)
