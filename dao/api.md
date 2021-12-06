# das_database server

resp common:
```json
{
  "err_no": 0,
  "err_msg": "",
  "data": {}
}
```

```go
// error code
const (
ApiCodeSuccess        ApiCode = 0
ApiCodeError500       ApiCode = 500
ApiCodeParamsInvalid  ApiCode = 10000
ApiCodeMethodNotExist ApiCode = 10001
ApiCodeDbError        ApiCode = 10002
ApiCodeCacheError     ApiCode = 10003
ApiCodeSystemUpgrade  ApiCode = 30019 
)
```

## If The Latest Block Number

* post: /isLatestBlockNumber
* req

```json
Nil
```

* resp
```json
{
  "apiCode": 0,
  "apiMsg": "",
  "data": {
    "isLatestBlockNumber": true
  }
}
```

## Parse Transaction

* post: /parserTransaction
* req

```json
{
  "txHash": "0xb940f751ac1d3293f6c47ce0a71e2712fd4931476bcff9339d16addda6dc99aa"
}
```

* resp

```json
{
  "apiCode": 0,
  "apiMsg": ""
}
```

## Api Test

```shell
curl -X POST http://127.0.0.1:8118/v1/latest/block/number

curl -X POST http://127.0.0.1:8118/v1/parser/transaction -d '{"txHash":"0x77a891bcec5b11d3fed14cfa5bd8cf5532f6d09cc6ecefa77d9e4bef296e8fd0"}'
```
