# tcp-proxy

Simple TCP proxy that was originally designed and used to explore redis RESP protocol. It may be useful for other scenarios as well.

E.g., Use it as the middle-man between redis server and redis-cli to log the transferred messages between the two.

```sh
# redis-server runnning on port 6379

# run tcp-proxy on 6378, and proxy traffic between the client and the redis-server
go run . -port=6378 -proxy-port=6379 -proxy-host=localhost

# client connects to 6378, the proxy
# redis-cli -p 6378
```

E.g., For MySQL

```sh
tcp-proxy -proxy-port=3306 -port=3307
```
