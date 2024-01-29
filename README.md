# Merkle Store

Merkle store is a server and client utility cmd used for uploading and downloading files backed by a Merkle tree.

## Demo

### server

The default server port is 3333.

```
cd cmd/client/server
go build .
./server
```

It can also be run with docker. The default docker-compose will boot two servers on port `3333` and `4444`. 

```
docker-compose up
```

### client

```
cd cmd/client/msc
go build .
./msc

./msc upload [FILES] --server SERVER_URL
./msc download ROOT_HASH [FILE_INDEXES] --server SERVER_URL
```

You can specify the server url with each command or put it in the env variable `MERKLE_STORE_SERVER`