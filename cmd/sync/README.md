# sync

## config

./conf.toml

```toml
infuraapi="https://ropsten.infura.io/v3/fa2b9f06ae2f43faa3ce69c80fde51c5"
mnemonic="large flower skin interest pulp embrace until jelly drop insane erase unveil"
to=["0x57C6829628b05EBE5B8ACF2116ba7334C28D6ceF","0x57C6829628b05EBE5B8ACF2116ba7334C28D6ceF"] #contract address
schedule="27 18 * * *" #CRON Expression Fromat  (https://godoc.org/github.com/robfig/cron)
```
## compile
cd sync
go build

## use
./sync