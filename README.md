# ALGOREP

## Build

`cd src`
`go build -o raft`

## Testing

`go get github.com/mattn/goreman`
`goreman start`

## Killing a node

`cd src_kill`
`go build -o raft_kill`
`./raft_kill address_of_nodes_to_kill`