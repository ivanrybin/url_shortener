# URL shortener

gRPC client and server with PostgreSQL database for URL shortening to 10 characters string.

# Usage example

Client has two commands `create <URL>` and `get <URL>`:

```bash
# create short URL from original
$ ./urls_client create google.com
3PjSsTTFog

# get original URL from shorted one
$ ./urls_client get 3PjSsTTFog
google.com 
```

# Build and deploy

### Server and database

```bash
$ sudo docker-compose -f docker-compose.yml up
```

Uploads and starts two docker images: [server](https://hub.docker.com/repository/docker/vnch/url_shortener_server)
and [PostgreSQL](https://hub.docker.com/repository/docker/vnch/url_shortener_db) database.

Server daemon waits PostgreSQL initialization finish.

```bash
# docker-compose log output after up command
pg_db       | 2021-08-30 10:00:46.295 UTC [1] LOG:  database system is ready to accept connections
urls_server | time="10:00:47 2021-08-30" level=info msg="trying to connect to database #1"
urls_server | time="10:00:47 2021-08-30" level=info msg="database connection established"
urls_server | time="10:00:47 2021-08-30" level=info msg="daemon started"
urls_server | time="10:00:47 2021-08-30" level=info msg="listening :9876"
```

After successful deployment you can request service by default port `9876` with gRPC client.

Server and database images setup available in `Dockerfile_server` and `Dockerfile_db` files.

You can build and run server locally without docker with your PostgresSQL database with schema from `db/create-table.sql`. 
```bash
$ ./build_server.sh
```

### Client

```bash
$ ./build_client.sh
```

Creates client binary file `urls_client`.

## Configuration

### Server and database

Server has config file (`configs/config.yml`) that mounts to server image in `docker-compose.yml`.

```yml
# configs/config.yml

server:
  host:            # default host
  port: 9876       # default port
  lru_size: 10000  # LRU cache size

database:
  host: database   # default db host in docker-compose
  port: 5432       # default db port in docker-compose

  name: docker     # \
  user: docker     #   db name and user info from db image `Dockerfile_db`
  password: docker # /

  conn_try_time: 5    # server db connection try duration  
  conn_tries_cnt: 10  # server db connection tries count  

  max_open_conns: 16  # golang database/sql driver
  max_idle_conns: 16  # golang database/sql driver

```

Database stores data in mounted directory `db/data`.

### Client

Client has one flag `-a, --address` for server address in format `localhost:9876` (default is `:9876`).

```bash
$ ./urls_client -a localhost:9876 create google.com
```

## gRPC protocol


```protobuf
// pkg/grpc/url_shortener.proto

syntax = "proto3";

service URLShortener {
  // shorts original URL and returns shorted URL
  rpc Create(CreateRequest) returns (CreateResponse) {};

  // returns original URL from shorted one
  rpc Get(GetRequest) returns (GetResponse) {};
}

message CreateRequest {
  string original_url = 1;
}

message CreateResponse {
  string short_url = 1;
}

message GetRequest {
  string short_url = 1;
}

message GetResponse {
  string original_url = 1;
}
```

## Hash algorithm

**Input**: arbitrary length string `input`.

**Output**: 10 chacharacters `hash` from alphabet `[_0-9a-zA-Z]`.

```go
alpha = '0123456789_ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz'
alpha_len = 63

sha  = sha256(input)   # == 32 bytes
sha  = sha[0:30]       # truncate 32 bytes to first 30 bytes

hash = string(10);     # 10 chars len string
 
for i = 0; i <= 27; i += 3 
    pos = to_integer(sha[i], sha[i+1], sha[i+2]) % alpha_len
    hash[i/3] = alpha[pos]
    
func to_integer(b1, b2, b3 byte) uint32:
    return uint32([0, b1, b2, b3])
```