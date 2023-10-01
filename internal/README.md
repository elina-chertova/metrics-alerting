# internal

В данной директории и её поддиректориях будет содержаться имплементация вашего сервиса

test

```azure
./metricstest-darwin-arm64 -test.v -test.run=^TestIteration1$  -binary-path=cmd/server/server

./metricstest-darwin-arm64 -test.v -test.run=^TestIteration2A$ -source-path=. -agent-binary-path=cmd/agent/agent
    
./metricstest-darwin-arm64 -test.v -test.run=^TestIteration2B$ -source-path=. -agent-binary-path=cmd/agent/agent
```