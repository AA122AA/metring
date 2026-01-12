build-server:
	/usr/local/go/bin/go build -o ./cmd/server/server ./cmd/server/main.go

build-agent:
	/usr/local/go/bin/go build -o ./cmd/agent/agent ./cmd/agent/main.go

run-server: build-server
	./cmd/server/server -a "localhost:8080" -i 10

run-agent: build-server
	./cmd/agent/agent -a "localhost:8080" -r 4

autotest-1: build-server build-agent
	./metricstest_v2 -test.v -test.count 1 -test.run=^TestIteration1$$ -binary-path=cmd/server/server

autotest-2: build-server build-agent
	./metricstest_v2 -test.v -test.run=^TestIteration2A$$ \
	-source-path=. \
	-agent-binary-path=cmd/agent/agent
	./metricstest_v2 -test.v -test.run=^TestIteration2B$$ \
	-source-path=. \
	-agent-binary-path=cmd/agent/agent

autotest-3: build-server build-agent
	./metricstest_v2 -test.v -test.run=^TestIteration3A$$ \
	-source-path=. \
	-agent-binary-path=cmd/agent/agent \
	-binary-path=cmd/server/server
	./metricstest_v2 -test.v -test.run=^TestIteration3B$$ \
	-source-path=. \
	-agent-binary-path=cmd/agent/agent \
	-binary-path=cmd/server/server

autotest-4: build-server build-agent
	./metricstest_v2 -test.v -test.run=^TestIteration4$$ \
	-agent-binary-path=cmd/agent/agent \
	-binary-path=cmd/server/server \
	-server-port=8080 \
	-source-path=.

autotest-5: build-server build-agent
	SERVER_PORT=8080 ADDRESS="localhost:$${SERVER_PORT}" TEMP_FILE="lol" ./metricstest_v2 \
	-test.v \
	-test.run=^TestIteration5$$ \
    -agent-binary-path=cmd/agent/agent \
    -binary-path=cmd/server/server \
    -server-port=8080 \
    -source-path=.

autotest-6: build-server build-agent
	./metricstest_v2 -test.v -test.run=^TestIteration6$$ \
	-agent-binary-path=cmd/agent/agent \
	-binary-path=cmd/server/server \
	-server-port=8080 \
	-source-path=.

autotest-7: build-server build-agent
	./metricstest_v2 -test.v -test.run=^TestIteration7$$ \
	-agent-binary-path=cmd/agent/agent \
	-binary-path=cmd/server/server \
	-server-port=8080 \
	-source-path=.

autotest-8: build-server build-agent
	./metricstest_v2 -test.v -test.run=^TestIteration8$$ \
	-agent-binary-path=cmd/agent/agent \
	-binary-path=cmd/server/server \
	-server-port=8080 \
	-source-path=.

autotest-9: build-server build-agent
	./metricstest_v2 -test.v -test.run=^TestIteration9$$ \
            -agent-binary-path=cmd/agent/agent \
            -binary-path=cmd/server/server \
            -file-storage-path="data/metrics.json" \
            -server-port=8080 \
            -source-path=.

autotests: build-server build-agent autotest-1 autotest-2 autotest-3 autotest-4 autotest-5 autotest-6 autotest-7 autotest-8 autotest-9
