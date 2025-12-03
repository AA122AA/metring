build:
	@go build -o bin/server/server ./cmd/server/main.go

run: build
	@./bin/server/server

autotest-1:
	./metricstest_v2 -test.v -test.run=^TestIteration1$$ -binary-path=cmd/server/server

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
