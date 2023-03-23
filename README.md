# das-database
A block parser tool that allows extraction of various data types on DAS
(register, edit, sell, transfer, ...) from CKB
## Prerequisites
* Ubuntu 18.04 or newer
* MYSQL >= 8.0
* go version >= 1.17.10
* [ckb-node](https://github.com/nervosnetwork/ckb) (Must be synced to latest height)
* [ckb-indexer](https://github.com/nervosnetwork/ckb-indexer) (Must be synced to latest height)
* If the version of the dependency package is too low, please install `gcc-multilib` (apt install gcc-multilib)

## Install & Run

### Source Compile
```bash
# get the code
git clone https://github.com/dotbitHQ/das-database.git

# init config/config.yaml
cp config/config.example.yaml config/config.yaml
 
# create mysql database
mysql -uroot -p
> create database das_database;
> quit;

# compile and run
cd das-database
make parser
./das_database_server --config=config/config.yaml
# it will take about 3 hours to synchronize to the latest data(Dec 6, 2021)
```

### Docker
* docker >= 20.10
* docker-compose >= 2.2.2


```bash
sudo curl -L "https://github.com/docker/compose/releases/download/v2.2.2/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose
sudo ln -s /usr/local/bin/docker-compose /usr/bin/docker-compose
docker-compose up -d
```

_if you already have a mysql installed, just run_
```bash
docker run -dp 8118:8118 -v $PWD/config/config.yaml:/app/config/config.yaml --name das-database-server dotbitteam/das-database:latest
```

## Usage
```sql
select * from das_database.t_account_info limit 10;
```

### Action Types
All supported parsable transaction types as following:

```txt
config              
deploy              
apply_register      
pre_register        
propose             
extend_proposal     
confirm_proposal    
edit_records        
edit_manager        
renew_account       
transfer_account    
withdraw_from_wallet
consolidate_income  
create_income       
transfer_balance    
start_account_sale  
edit_account_sale   
cancel_account_sale 
buy_account         
```

### Tables

* t_account_info
* t_trade_info
* t_income_cell_info
* t_block_info (Only store the latest 20 blocks in case of rollback)
* t_trade_deal_info
* t_rebate_info (Records of inviter/channel's rewards)
* t_records_info
* t_token_price_info
* t_transaction_info 
* t_reverse_records_info (All transactions on DAS)

More details see [das_database.sql](https://github.com/dotbitHQ/das-database/blob/main/dao/das_database.sql)

## Others
* [What is DAS](https://github.com/dotbitHQ/das-contracts/blob/master/docs/en/Overview-of-DAS.md)
* [What is a DAS transaction on CKB](https://github.com/dotbitHQ/das-contracts/blob/master/docs/en/Data-Structure-and-Protocol/Transaction-Structure.md)
