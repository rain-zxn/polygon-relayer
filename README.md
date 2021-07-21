Polygon Relayer is an important character of Poly cross-chain interactive protocol which is responsible for relaying cross-chain transaction from and to Polygon.

## Build From Source

### Prerequisites

- [Golang](https://golang.org/doc/install) version 1.14 or later

### Build

```shell
git clone https://github.com/polynetwork/polygon-relayer.git
cd polygon-relayer
CGO_ENABLED=0 go build -o polygon-relayer main.go
```

After building the source code successfully,  you should see the executable program `polygon-relayer`. 

### Mock integrated test
```shell
./polygon-relayer --cliconfig ./env/prod/config.json --testlocal
```
### Build Docker Image

```
docker build -t polynetwork/polygon-relayer -f Dockerfile ./
```

This command will copy ./config.json to /app/config.json in the image. So you need to prepare config.json before running this command and you should start the polygon-relayer in container basing on the configuration in /app/config.json.

## Run Relayer

Before you can run the relayer you will need to create a wallet file of PolyNetwork. After creation, you need to register it as a Relayer to Poly net and get consensus nodes approving your registeration. And then you can send transaction to Poly net and start relaying.

Before running, you need feed the configuration file `config.json`.

```
{
  "PolyConfig":{
    "RestURL":"http://poly_ip:20336", // address of Poly
    "EntranceContractAddress":"0300000000000000000000000000000000000000", // CrossChainManagerContractAddress on Poly. No need to change
    "WalletFile":"./wallet.dat", // your poly wallet
    "WalletPwd":"pwd" //password
  },
  "ETHConfig":{
    "SideChainId": 2, // polygon chainID
    "RestURL":"http://polygon:port", // your polygon node 
    "ECCMContractAddress":"polygon_cross_chain_contract", 
    "ECCDContractAddress":"polygon_cross_chain_data_contract",
    "KeyStorePath": "./keystore", // path to store your polygon wallet
    "KeyStorePwdSet": { // password to protect your polygon wallet
      "0xd12e...54ccacf91ca364d": "pwd1", // password for address "0xd12e...54ccacf91ca364d"
      "0xabb4...0aba7cf3ee3b953": "pwd2" // password for address "0xabb4...0aba7cf3ee3b953"
    },
    "BlockConfig": 12, // blocks to confirm a polygon tx
    "HeadersPerBatch": 500, // number of poly headers commited to ECCM in one transaction at most
    "MonitorInterval": 3 // seconds of ticker to monitor polygon chain
  },
  "BoltDbPath": "./db", // DB path
  "RoutineNum": 64,
  "TargetContracts": [
    {
      "0xD8aE73e06552E...bcAbf9277a1aac99": { // your lockproxy hash
        "inbound": [6], // from which chain allowed
        "outbound": [6] // to which chain allowed
      }
    }
  ]
}
```

After that, make sure you already have a polygon wallet with MATIC. The wallet file is like `UTC--2020-08-17T03-44-00.191825735Z--0xd12e...54ccacf91ca364d` and you can use [geth](https://github.com/ethereum/go-ethereum) to create one( `./geth account new --datadir .` ). Put it under `KeyStorePath`. You can create more than one wallet for relayer. Relayer will send transactions concurrently by different accounts.

Now, you can start relayer as follow: 

```shell
./eth_relayer --cliconfig=./config.json 
```

It will generate logs under `./Log` and check relayer status by view log file.


