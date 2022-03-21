# Harmony Analytics Toolkit 
(internally harmolytics)

-----

This project is very much so build by me for me. Right now it doesn't support anything particularly useful and isn't necessarily made with other users in mind.

Nonetheless, if you want to use it or build upon it feel free to test it. You will need a MySQL instance as a backend service.

## Setup
### Manual Setup
If you have go installed and access to your own MySQL DB you can simply build the tools and use it as is 
```
git clone https://github.com/mjmar01/harmolytics.git
cd harmolytics
go build cmd/harmolytics
go build cmd/harmony-tool
```
Configure the database. Enter password and create keyring when prompted
```
./harmolytics config database --db-host <host> --db-port <port> --db-user <user>
./harmolytics sql init
```
### Docker Setup
For this you need nothing other than docker and docker-compose installed. 
This setup will create a temporary MySQL DB and build all tools for you
```
git clone https://github.com/mjmar01/harmolytics.git
cd harmolytics/build
./build.sh
```
Configure the database. The default password is `mypassword`. Create keyring if prompted.
```
./harmolytics config database --db-host 127.0.0.1 --db-port 3306 --db-user root
./harmolytics sql init
```
## Basic start
```
# Load raw transactions (by wallet or txHashs)
./harmolytics load transactions -a <one1... 0x...>
./harmolytics load transactions <txHash> <txHash>...
# Load raw information about used tokens and methods
./harmolytics load methods
./harmolytics load tokens

# Process previously loaded raw information into a more readable format 
./harmolytics decode swaps
./harmolytics decode transfers
./harmolytics decode liquidity-pools
./harmolytics decode liquidity-actions

# After identifying liquidity-pools load historic ratios to calculate prices
./harmolytics load ratios

# Analyze stored data
./harmolytics analyze fees
```
If nothing ended up on fire by now you should have a database full of all kinds of data. For a better overview there are views however I recommend using raw data until displaying to avoid precision errors.

## Known problems
There isn't a lot of input validation, so you could easily break it if you wanted to.
There is also an SQL injection vulnerability in the profile parameter.

## Todo
Quite a bit I still want to do with this. No guarantee any of this will ever make it.

- Historical price data
- Better info regarding masterChef contracts and similar
- Fee analysis (~~Uniswap~~, Tokenomics, etc...)
- Some sort of visualization tools like portfolio over time etc...
- More resilient design and error handling
- CSV export capabilities
- Documentation, better help commands, input validation and stuff
- Clean harmony, rpc, and solidityio package to allow external use

## Contribution
As said above this is more of a personal project so no need for anyone else to deal with this. However, if you have feedback or improvement ideas feel free to contact me over this repo or on harmony discord @Markus#8518
