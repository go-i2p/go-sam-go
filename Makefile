fmt:
	find . -name '*.go' -exec gofumpt -w -s -extra {} \;

export DEBUG_I2P=debug
#export WARNFAIL_I2P=true
common-test:
	go test --tags nettest -v ./common/...

datagram-test:
	go test --tags nettest -v ./datagram/...

primary-test:
	go test --tags nettest -v ./primary/...

stream-test:
	go test --tags nettest -v ./stream/...

raw-test:
	go test --tags nettest -v ./raw/...

test:
	make common-test
	make stream-test
	make datagram-test
	#make raw-test
	#make primary-test


test-logs:
	make common-test 2> common-err.log 1> common-out.log;	cat common-err.log common-out.log
	make stream-test 2> stream-err.log 1> stream-out.log;	cat stream-err.log stream-out.log
	make datagram-test 2> datagram-err.log 1> datagram-out.log;	cat datagram-err.log datagram-out.log
	#make raw-test 2> raw-err.log 1> raw-out.log;	cat raw-err.log raw-out.log
	#make primary-test 2> primary-err.log 1> primary-out.log;	cat primary-err.log primary-out.log
