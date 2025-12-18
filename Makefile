build:
	@go build -o bin/server/server ./cmd/server/main.go

run: build
	@./bin/server/server

autotest-1:
	./metricstest_v2 -test.v -test.count 1 -test.run=^TestIteration1$$ -binary-path=cmd/server/server

autotest-2:
	./metricstest_v2 -test.v -test.run=^TestIteration2A$$ \
	-source-path=. \
	-agent-binary-path=cmd/agent/agent
	./metricstest_v2 -test.v -test.run=^TestIteration2B$$ \
	-source-path=. \
	-agent-binary-path=cmd/agent/agent

autotest-3:
	./metricstest_v2 -test.v -test.run=^TestIteration3A$$ \
	-source-path=. \
	-agent-binary-path=cmd/agent/agent \
	-binary-path=cmd/server/server
	./metricstest_v2 -test.v -test.run=^TestIteration3B$$ \
	-source-path=. \
	-agent-binary-path=cmd/agent/agent \
	-binary-path=cmd/server/server

autotest-4:
	./metricstest_v2 -test.v -test.run=^TestIteration4$$ \
	-agent-binary-path=cmd/agent/agent \
	-binary-path=cmd/server/server \
	-server-port=8080 \
	-source-path=.

autotest-5:
	SERVER_PORT=8080 ADDRESS="localhost:$${SERVER_PORT}" TEMP_FILE="lol" ./metricstest_v2 \
	-test.v \
	-test.run=^TestIteration5$$ \
    -agent-binary-path=cmd/agent/agent \
    -binary-path=cmd/server/server \
    -server-port=8080 \
    -source-path=.

autotest-6:
	./metricstest_v2 -test.v -test.run=^TestIteration6$$ \
	-agent-binary-path=cmd/agent/agent \
	-binary-path=cmd/server/server \
	-server-port=8080 \
	-source-path=.

autotest-7:
	./metricstest_v2 -test.v -test.run=^TestIteration7$$ \
	-agent-binary-path=cmd/agent/agent \
	-binary-path=cmd/server/server \
	-server-port=8080 \
	-source-path=.
