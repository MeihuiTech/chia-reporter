##Development
### Requirements

mysql 5.7 above

### Run with docker

```
docker build -t chia-block-sync:VERSION .

docker run -ti --name chia-block-sync -v $PATH_TO_CONFIG/config.json:/go/src/app/config.json $PATH_TO_CERTS:/go/src/app/certs chia-block-sync:VERSION
```

### Configuration

#### Config example
```
{
  "dsn": "USERNAME:PASSWORD@tcp(DB_HOST:DB_PORT)/DB_NAME?charset=utf8mb4&parseTime=True&loc=Local",
  "rpc_host": "CHIA_RCP_HOST",
  "rpc_port": CHIA_FULL_NODE_RPC_PORT,
  "private_cert": "PATH_TO_PRIVATE_FULL_NODE.CRT",
  "private_key": "PATH_TO_PRIVATE_FULL_NODE.KEY",
  "ca_cert": "PATH_TO_PRIVATE_CA.CRT"
}
```

#### variables
- USERNAME 
    
    username of the database to store blockchain data
- PASSWORD 

    password of the database
- DB_HOST
    
    host of the database
- DB_PORT
    
    port of the database
- DB_NAME
    
    name of the schema
- CHIA_RPC_HOST 
    
    host of the chia full node without the schema eg: 192.168.0.111
- CHIA_FULL_NODE_RCP_PORT

    port of the chia full node
- PATH_TO_PRIVATE_FULL_NODE.CRT
    
    chia full node's private cert which can be find at chia full node's dir `~/.chia/mainnet/config/ssl/full_node/private_full_node.cert`
    
- PATH_TO_PRIVATE_FULL_NODE.KEY
    
    chia full node's private cert which can be find at chia full node's dir `~/.chia/mainnet/config/ssl/full_node/private_full_node.key`
    
- PATH_TO_PRIVATE_CA.CRT
    
    chia full node's private cert which can be find at chia full node's dir `~/.chia/mainnet/config/ssl/ca/private_ca.crt`



