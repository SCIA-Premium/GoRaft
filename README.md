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

The objective of this project is to set up a **client/server** system with a mechanism to controll or inject faults into the system.

The general idea is as follows: customers offer values / commands to **servers**. These servers then want to agree on the order in which they will accept, then run these commands. Once agreed, it will write them to a **log file**, and execute them. 

---

## RAFT

### Build

`cd src`
`go build -o raft`

### Testing

`go get github.com/mattn/goreman`
`goreman start`

---

## REPL

`cd src_repl`
`go build -o raft_repl`

`./raft_repl --to_kill address_of_nodes`
`./raft_repl --to_speed [low|medium|high],address_of_nodes`

---

## CLSF