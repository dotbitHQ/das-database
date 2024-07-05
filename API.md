* [API List](#API-List)
    * [Get Account History Permission](#Get-Account-History-Permission)
    * [Get Address History Hold Accounts](#Get-Address-History-Hold-Accounts)
    * [Get Snapshot Register Progress](Get-Snapshot-Register-Progress)
    * [Get Snapshot Progress](#Get-Snapshot-Progress)

## API List

Please familiarize yourself with the meaning of some common parameters before reading the API list:

| param                                                                          | description                                        |
|:-------------------------------------------------------------------------------|:---------------------------------------------------|
| type                                                                           | Filled with "blockchain" for now                   |
| [coin_type](https://github.com/satoshilabs/slips/blob/master/slip-0044.md)     | 60: eth, 195: trx, 9006: bsc, 966: matic, 3: doge  |
| account                                                                        | Contains the suffix `.bit` in it                   |
| key                                                                            | Generally refers to the blockchain address for now |

### Error Code

```txt
ApiCodeError500       ApiCode = 500 // server err
ApiCodeParamsInvalid  ApiCode = 10000 // invalid params
ApiCodeMethodNotExist ApiCode = 10001 // method not exist
ApiCodeDbError        ApiCode = 10002 // db err
ApiCodeAccountPermissionsDoNotExist ApiCode = 30020 // Account permission does not exist
ApiCodeAccountHasBeenRecycled       ApiCode = 30021 // Account has been recycled
ApiCodeAccountCrossChain            ApiCode = 30022 // Account cross-chain
ApiCodeAccountExpired               ApiCode = 20023 // account expired

```

### Algorithm ID

```txt

DasAlgorithmIdEth       DasAlgorithmId = 3 // ETH Personal Sign
DasAlgorithmIdTron      DasAlgorithmId = 4 // TRON Personal Sign
DasAlgorithmIdEth712    DasAlgorithmId = 5 // ETH 712 Sign
DasAlgorithmIdEd25519   DasAlgorithmId = 6 // Ed25519
DasAlgorithmIdDogeChain DasAlgorithmId = 7 // Doge Sign
```

### Get Account History Permission

**Request**
* path: /v1/snapshot/permissions/info
* param:

```json
{
  "account": "7aaaaaaa.bit",
  "block_number": 3593828
}
```

**Response**

* owner_algorithm_id: The algorithm flag of the chain to which the address belongs

```json
{
  "errno": 0,
  "errmsg": "",
  "data": {
    "account": "7aaaaaaa.bit",
    "account_id": "0xc475fcded6955abc8bf6e2f23e68c6912159505d",
    "block_number": 3512736,
    "owner": "0xc9f53b1d85356b60453f867610888d89a0b667ad",
    "owner_algorithm_id": 5,
    "manager": "0xc9f53b1d85356b60453f867610888d89a0b667ad",
    "manager_algorithm_id": 5
  }
}
```

**Usage**

```shell
curl -X POST http://127.0.0.1:8118/v1/snapshot/permissions/info -d'{"account":"7aaaaaaa.bit","block_number":3593828}'
```

or json rpc style:

```shell
curl -X POST http://127.0.0.1:8118 -d'{"jsonrpc": "2.0","id": 1,"method": "snapshot_permissions_info","params": [{"account":"7aaaaaaa.bit","block_number":3593828}]}'
```

### Get Address History Hold Accounts

**Request**
* path: /v1/snapshot/address/accounts
* param:
    * role_type: (permission role type) manager or owner
    * size: [1,100]
```json
{
  "type": "blockchain",
  "key_info": {
    "coin_type": "195",
    "key": "41a2ac25bf43680c05abe82c7b1bcc1a779cff8d5d"
  },
  "block_number": 1941502,
  "role_type": "manager",
  "page": 1,
  "size": 100
}
```

**Response**

```json
{
  "errno": 0,
  "errmsg": "",
  "data": {
    "accounts": [
      {
        "account": "8aaaaaaa.bit"
      },
      {
        "account": "9aaaaaaa.bit"
      }
    ]
  }
}
```

**Usage**

```shell
curl -X POST http://127.0.0.1:8118/v1/snapshot/address/accounts -d'{"type":"blockchain","key_info":{"coin_type":"195","key":"41a2ac25bf43680c05abe82c7b1bcc1a779cff8d5d"},"block_number":1941502,"role_type":"manager"}'
```

or json rpc style:

```shell
curl -X POST http://127.0.0.1:8118 -d'{"jsonrpc": "2.0","id": 1,"method": "snapshot_address_accounts","params": [{"type":"blockchain","key_info":{"coin_type":"195","key":"41a2ac25bf43680c05abe82c7b1bcc1a779cff8d5d"},"block_number":1941502,"role_type":"manager"}]}'
```

### Get Snapshot Progress

**Request**
* path: /v1/snapshot/progress
* param:
  * role_type: (permission role type) manager or owner
  * size: [1,100]
```json
{
  "block_number": 1941502
}
```

**Response**

```json
{
  "errno": 0,
  "errmsg": "",
  "data": {
    "block_number": 1941502
  }
}
```

**Usage**

```shell
curl -X POST http://127.0.0.1:8118/v1/snapshot/progress -d'{"block_number": 1941502}'
```

### Get Snapshot Register Progress

**Request**
* path: /v1/snapshot/register/history
* param:
  * role_type: (permission role type) manager or owner
  * size: [1,100]
```json
{
  "start_time": 0
}
```

**Response**

```json
{
  "errno": 0,
  "errmsg": "",
  "data": {
    "result": ""
  }
}
```

**Usage**

```shell
curl -X POST http://127.0.0.1:8118/v1/snapshot/register/history -d'{"start_time": 0}'
```
