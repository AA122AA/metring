PORT ?= 8082

DB_USER ?= metring
DB_PASSWORD ?= StrongPass123!
DB_NAME ?= metring

DSN ?= "postgresql://${DB_USER}:${DB_PASSWORD}@localhost:5432/${DB_NAME}?sslmode=disable"

##########################################################################################

## >>> Developer <<<

build-server:
	/usr/local/go/bin/go build -o ./cmd/server/server ./cmd/server/main.go

build-agent:
	/usr/local/go/bin/go build -o ./cmd/agent/agent ./cmd/agent/main.go

run-server: build-server
	./cmd/server/server -a "localhost:${PORT}" -i 10

run-server-file: build-server
	./cmd/server/server -a "localhost:${PORT}" -i 10 -f "data/metrics.json"

run-server-db: build-server
	./cmd/server/server -a "localhost:${PORT}" -i 10 -d ${DSN}

run-agent: build-agent
	./cmd/agent/agent -a "localhost:${PORT}" -r 4

tidy:
	/usr/local/go/bin/go mod tidy && /usr/local/go/bin/go mod vendor

goose-create-init:
	goose postgres ${DSN} -s -dir ./db/schema create init sql

goose-status:
	goose postgres ${DSN} -dir ./db/schema status

goose-up:
	goose postgres ${DSN} -dir ./db/schema up

##########################################################################################

## >>> Database <<<
select-all:
	PGPASSWORD=${DB_PASSWORD} psql -h localhost -U ${DB_USER} -c "SELECT * from metrics ORDER BY name;"

delete-all:
	PGPASSWORD=${DB_PASSWORD} psql -h localhost -U ${DB_USER} -c "TRUNCATE TABLE metrics RESTART IDENTITY;"

##########################################################################################

## >>> Autotests <<<

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

autotest-10: build-server build-agent
	./metricstest_v2 -test.v -test.run=^TestIteration10A$$ \
            -agent-binary-path=cmd/agent/agent \
            -binary-path=cmd/server/server \
            -file-storage-path="data/metrics.json" \
            -database-dsn=${DSN} \
            -server-port=8080 \
            -source-path=.
	./metricstest_v2 -test.v -test.run=^TestIteration10B$$ \
	        -agent-binary-path=cmd/agent/agent \
			-binary-path=cmd/server/server \
            -file-storage-path="data/metrics.json" \
            -database-dsn=${DSN} \
            -server-port=8080 \
            -source-path=.

autotest-11: build-server build-agent
	./metricstest_v2 -test.v -test.run=^TestIteration11$$ \
            -agent-binary-path=cmd/agent/agent \
            -binary-path=cmd/server/server \
            -database-dsn=${DSN} \
            -server-port=8080 \
            -source-path=.

autotest-12: build-server build-agent
	./metricstest_v2 -test.v -test.run=^TestIteration12$$ \
            -agent-binary-path=cmd/agent/agent \
            -binary-path=cmd/server/server \
            -database-dsn=${DSN} \
            -server-port=8080 \
            -source-path=.

autotests: build-server build-agent autotest-1 autotest-2 autotest-3 autotest-4 autotest-5 autotest-6 autotest-7 autotest-8 autotest-9 autotest-10 autotest-11 autotest-12
