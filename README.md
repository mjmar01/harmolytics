# Harmony Analytics Toolkit 
(internally harmolytics)

-----

This project is very much so build by me for me. Right now it doesn't support anything particularly useful and isn't necessarily made with other users in mind.

Nonetheless, if you want to use it or build upon it feel free to test it. You will need a MySQL instance as a backend service.

## Basic start
```
./harmolytics config database --db-host 127.0.0.1 --db-port 3306 --db-user <user>
./harmolytics sql init

./harmolytics load transactions -a <one1... 0x...>
./harmolytics load transactions <txHash> <txHash>...
./harmolytics load methods
./harmolytics load tokens

./harmolytics decode swaps
./harmolytics decode liquidity-actions
./harmolytics decode transfers
./harmolytics decode liquidity-pools

./harmolytics load ratios
```
If nothing ended up on fire by now you should have a database full of all kinds of data. For a better overview there are views however I recommend using raw data until displaying to avoid precision errors.

## Known problems
There isn't a lot of input validation, so you could easily break it if you wanted to.
There is also an SQL injection vulnerability in the profile parameter. 
Password safety isn't top-notch either. Again this is mostly made by myself for myself.

## Todo
Quite a bit I still want to do with this. No guarantee any of this will ever make it.

- Historical price data
- Better info regarding masterChef contracts and similar
- Fee analysis (Uniswap, Tokenomics, etc...)
- Some sort of visualization tools like portfolio over time etc...
- More resilient design and error handling
- CSV export capabilities and maybe even remove the need for a database
- Documentation, better help commands, input validation and stuff

## Contribution
As said above this is more of a personal project so no need for anyone else to deal with this. However, if you have feedback or improvement ideas feel free to contact me over this repo or on harmony discord @Markus#8518
