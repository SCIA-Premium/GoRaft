# ALGOREP

## Build

`cd src`
`go build -o raft`

## Testing

`go get github.com/mattn/goreman`
`goreman start`

## REPL

`cd src_repl`
`go build -o raft_repl`

`./raft_repl --to_kill address_of_nodes`
`./raft_repl --to_speed [low|medium|high],address_of_nodes`
