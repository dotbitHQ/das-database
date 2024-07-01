# das-database
A block parser tool that allows extraction of various data types on DAS
(register, edit, sell, transfer, ...) from CKB
## Prerequisites
* Ubuntu 18.04 or newer
* MYSQL >= 8.0
* go version >= 1.21.3
* [ckb-node](https://github.com/nervosnetwork/ckb) (Must be synced to latest height and add `Indexer` module to ckb.toml)
* If the version of the dependency package is too low, please install `gcc-multilib` (apt install gcc-multilib)
* Machine configuration: 4c8g200G

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
make default
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
docker run -dp 8118:8118 -v $PWD/config/config.yaml:/app/config/config.yaml --name das-database-server admindid/das-database:latest
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
refund_apply
pre_register        
propose             
extend_proposal     
confirm_proposal   
recycle_proposal 
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
declare_reverse_record
redeclare_reverse_record
retract_reverse_record
transfer
create_approval
delay_approval
revoke_approval
fulfill_approval
make_offer
edit_offer
cancel_offer
accept_offer
enable_sub_account
create_sub_account
edit_sub_account
renew_sub_account
recycle_sub_account
lock_sub_account_for_cross_chain
unlock_sub_account_for_cross_chain
config_sub_account_custom_script
config_sub_account
collect_sub_account_profit
update_sub_account
collect_sub_account_channel_profit
lock_account_for_cross_chain
unlock_account_for_cross_chain
force_recover_account_status
recycle_expired_account
update_reverse_record_root
create_device_key_list
update_device_key_list
mint_dp
transfer_dp
burn_dp
bid_expired_account_dutch_auction
account_cell_upgrade
did_edit_owner
did_edit_records
did_renew
did_recycle
did_upgrade
did_register
did_auction
```

## Others
* [What is DAS](https://github.com/dotbitHQ/did-contracts/blob/docs/docs/en/design/Overview-of-DAS.md)
* [What is a DAS transaction on CKB](https://github.com/dotbitHQ/did-contracts/blob/docs/docs/en/developer/Transaction-Structure.md)
