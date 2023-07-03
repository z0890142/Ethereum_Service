## 系統架構圖
```mermaid

graph LR

RPCEndpoint[RPCEndpoint]

subgraph indexer_service
    scan[scanHandler]
    scanner1[block_scanner]
    scanner2[tx_scanner]
    scanner3[log_scanner]

    consumer1[block_consumer]
    consumer2[tx_consumer]
    consumer3[log_consumer]
end

subgraph api_service
    dataHandler1[mysql_handler]
    dataHandler2[redis_handler]

    scanner4[block_scanner]
    scanner5[tx_scanner]
    scanner6[log_scanner]
end

service1[producer_service]

MQ[RabbitMQ]

Redis1[Redis]

Mysql1[Mysql]

service1 --> MQ
MQ --> scan
scan --> scanner1
scan --> scanner2
scan --> scanner3
scanner1 --> consumer1
scanner2 --> consumer2
scanner3 --> consumer3

consumer1 --> Mysql1
consumer2 --> Mysql1
consumer3 --> Mysql1

dataHandler1 --> Mysql1
dataHandler2 --> Redis1

scanner1 --> RPCEndpoint
scanner2 --> RPCEndpoint
scanner3 --> RPCEndpoint
scanner4 --> RPCEndpoint
scanner5 --> RPCEndpoint
scanner6 --> RPCEndpoint
service1 --> RPCEndpoint

```

### 說明
主要服務會分為三塊
* producer
    控制還有哪些 block 沒有被掃描進 database，會持續將需要掃描的 block number 送至 message queue 中，並且透過 message queue 得知已掃描的 block。

* indexer_service
    從啟動數個 worker 各自讀取 message queue 並進行 block 掃描，掃描並儲存至 database 後將透過 message queue 告知 producer
* api_service
    優先讀取 Redis 與 Database 內的資料，如 Redis 與 Database 內皆無資料，會向 RPC Endpoint 發起 Request 索取需要資料。

###啟動方式
```
make eth_service
```
---

### Config
```
ENV: local
SERVICE:
  NAME: api_service
  HOST: "0.0.0.0"
  PORT: "80"

LOG_LEVEL : INFO
LOG_FILE: stdout

DATABASES:
  DRIVER: mysql
  HOST: mysql
  PORT: 3306
  USERNAME: root
  PASSWORD: pass
  DBNAME: eth_service
  CHARSET: utf8mb4
  POOL_SIZE: 100
  TIMEOUT: 1s
  READ_TIMEOUT: 1s
  WRITE_TIMEOUT: 1s

REDIS:
  HOST: redis
  PORT: 6379
  PASSWORD: pass

STORE_BUFFER_SIZE: 1000
MIGRATION_FILE_PATH: ../../pkg/database/migrations

MAX_RETRY_TIME: 10
RCP_ENDPOINT: https://data-seed-prebsc-2-s3.binance.org:8545/
WORKER_NUMBER: 100
MQ_ENDPOINT: amqp://user:pass@rabbitmq:5672/
```

* DATABASES :
    mysql 相關設定
* REDIS :
    redis 相關設定
* STORE_BUFFER_SIZE :
    indexer_service 從 message queue 先拿回來的資料暫存數量
* WORKER_NUMBER :
    worker 數量
