# das-database

## Prerequisites
You will need
* Ubuntu 18.04 or newer
* MYSQL >= 8.0
* go version >= 1.15.0

to build and run das-database.

## Install
```bash
# get the code
git clone https://github.com/DeAccountSystems/das-database.git

# edit config/config.yaml before init mysql database
mysql -uroot -p
> source das-database/dao/das_database.sql
> quit;

# compile and run
cd das-database
make parser
./das_database_server --config=config/config.yaml
# it will take about 3 hours to synchronize to the latest data(Dec 6, 2021)
```
## Usage
Normal sql usage:
```sql
select * from das_database.t_account_info limit 10;
```

## Action Types
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

## Tables

* t_account_info
* t_trade_info
* t_income_cell_info
* t_block_info // Only store the latest 20 blocks in case of rollback
* t_trade_deal_info
* t_rebate_info // Records of inviter/channel's rewards
* t_records_info
* t_token_price_info
* t_transaction_info 
* t_reverse_records_info // All transactions on DAS

More details see [das_database.sql](https://github.com/DeAccountSystems/das-database/blob/main/dao/das_database.sql)
