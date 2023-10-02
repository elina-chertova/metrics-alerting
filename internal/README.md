# internal

В данной директории и её поддиректориях будет содержаться имплементация вашего сервиса

test

```azure
./metricstest-darwin-arm64 -test.v -test.run=^TestIteration1$  -binary-path=cmd/server/server

./metricstest-darwin-arm64 -test.v -test.run=^TestIteration2A$ -source-path=. -agent-binary-path=cmd/agent/agent
    
./metricstest-darwin-arm64 -test.v -test.run=^TestIteration2B$ -source-path=. -agent-binary-path=cmd/agent/agent
    
    
SERVER_PORT=90999 ADDRESS="localhost:${SERVER_PORT}" TEMP_FILE='temp'
./metricstest-darwin-arm64 -test.v -test.run=^TestIteration8$ \
            -agent-binary-path=cmd/agent/agent \
            -binary-path=cmd/server/server \
            -server-port=$SERVER_PORT \
            -source-path=.
    
SERVER_PORT=8080 ADDRESS="localhost:${SERVER_PORT}" 
./metricstest-darwin-arm64 -test.v -test.run=^TestIteration8$ \
            -agent-binary-path=cmd/agent/agent \
            -binary-path=cmd/server/server \
            -server-port=$SERVER_PORT \
            -source-path=.
```