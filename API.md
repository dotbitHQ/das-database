* [API List](#api-list)
    * [Get Account History Permissions](#Get Account History Permissions)

## API List

Please familiarize yourself with the meaning of some common parameters before reading the API list:

| param                                                                                    | description                                                         |
| :-------------------------                                                               | :------------------------------------------------------------------ |
| type                                                                                     | Filled with "blockchain" for now                                    |
| coin\_type <sup>[1](https://github.com/satoshilabs/slips/blob/master/slip-0044.md)</sup> | 60: eth, 195: trx, 9006: bsc, 966: matic                             |
| chain\_id <sup>[2](https://github.com/ethereum-lists/chains)</sup>                       | 1: eth, 56: bsc, 137: polygon; 5: goerli, 97: bsct, 80001: mumbai   |
| account                                                                                  | Contains the suffix `.bit` in it                                    |
| key                                                                                      | Generally refers to the blockchain address for now                  |

_You can provide either `coin_type` or `chain_id`. The `coin_type` will be used, if you provide both._

### Get Account History Permissions

**Request**

* param: none

```json
{
  "account": "7aaaaaaa.bit",
  "block_number": 3593828
}
```

**Response**

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