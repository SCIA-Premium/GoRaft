all: raft raft_repl raft_clsf

raft:
	cd src; go build -o raft; mv raft ..

raft_repl:
	cd src_repl; go build -o raft_repl; mv raft_repl ..

raft_clsf:
	cd src_clsf; go build -o raft_clsf; mv raft_clsf ..

clean:
	$(RM) -r raft raft_repl raft_clsf output

.PHONY: raft raft_repl raft_clsf