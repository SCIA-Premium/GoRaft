# ALGOREP [![Profile][title-img]][profile]

[title-img]:https://img.shields.io/badge/-SCIA--PRIME-red
[profile]:https://github.com/Pypearl

## AUTHORS
Aeden Bourgue \<aeden.bourgue@epita.fr\> \
Alexandre Lemonnier \<alexandre.lemonnier@epita.fr\>\
Eliott Bouhana \<eliott.bouhana@epita.fr\> \
Philippe Bernet \<philippe.bernet@epita.fr\> \
Sarah Gutierez \<sarah.gutierez@epita.fr\> \
Victor Simonin \<victor.simonin@epita.fr\>

---

The objective of this project is to set up a **client/server** system with a
mechanism to control or inject faults into the system.

The general idea is as follows: customers offer values / commands to
**servers**. These servers then want to agree on the order in which they will
accept, then run these commands. Once agreed, it will write them to a **log
file**, and execute them. 

---

## Build

The provided `Makefile` build all the necessary executables with the following
simple command : 

```bash
make
```

It is also possible to build the executable separately :

```bash
make raft
make raft_repl
make raft_clsf
```

## RAFT

### Testing

```bash
go get github.com/mattn/goreman
goreman start
```

Procfile format :

```bash
nodename: ./raft --peer_id NUMBER --peer PEER_ADDRESSES --port :PORT
```

As an example :
```bash
node1: ./raft --peer_id 1 --peer 127.0.0.1:22379,127.0.0.1:32379 --port 12379
node2: ./raft --peer_id 2 --peer 127.0.0.1:12379,127.0.0.1:32379 --port 22379
node3: ./raft --peer_id 3 --peer 127.0.0.1:12379,127.0.0.1:22379 --port 32379
```

---

## REPL

### Testing

```bash
./raft_repl address START
./raft_repl address CRASH
./raft_repl address SPEED (low|medium|high)
./raft_repl address RECOVERY
```

---

## CLSF

### Testing

```bash
./raft_clsf address LOAD filename
./raft_clsf address DELETE uuid
./raft_clsf address LIST
./raft_clsf address APPEND uuid content
```