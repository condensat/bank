# Condensat Bank backend

This repository hold all the backend components for Condensat Bank.

## Logging system

### Start mariadb

Log can be stored into mariadb.

```bash
docker run --name mariadb-test -e MYSQL_RANDOM_ROOT_PASSWORD=yes -e MYSQL_USER=condensat -e MYSQL_PASSWORD=condensat -e MYSQL_DATABASE=condensat -v $(pwd)/tests/database/permissions.sql:/docker-entrypoint-initdb.d/permissions.sql:ro -p 3306:3306 -d mariadb:10.3
```

### Start redis

Redis is used as a cache for logging to avoid message loses.

``` bash
docker run --name redis-test -p 6379:6379 -d redis:5-alpine
```

## Messaging system

Nats is used for internal messaging system between components.

### Start nats

``` bash
docker run --name nats-test -p 4222:4222 -d nats:2.1-alpine
```

### Start the log grabber
The log grabber fetch log entries from redis and display them.
Log entries are remove from redis after store


```bash
go run logger/cmd/grabber/main.go --log=debug > ../debug.log
```

### Start the log grabber with database
The log grabber fetch log entries from redis and store them to database.
Log entries are remove from redis after store


```bash
go run logger/cmd/grabber/main.go --log=debug --withDatabase=true
```

### Use RedisLogger

A logging component setup a RedisLogger and log normally.

```bash
go run logger/cmd/example/main.go --appName=Foo --log=debug
```

## Modify `/etc/hosts`

To run the bank stack locally, we need to add the following lines to the `/etc/hosts` file:
```
127.0.0.1       nats nats-test
127.0.0.1       redis redis-test cache
127.0.0.1       mariadb mariadb-test
127.0.0.1       db db-test
```

## Modify the `.gitconfig` file

To prevent an error with the `secureid` module, we need to add the following lines to `$HOME/.gitconfig`:
```
[url "git@code.condensat.tech:2222"]
        insteadOf = https://code.condensat.tech/
```

## Unit testing

```bash
go test -v ./...
```

## Create secureid.json file

We need a `secureid` file for the api service. Here's an example:
```
{
    "seed": "1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa", // use a base58 string
    "context": "BankApi",
    "keyId": 1
}
```

This file must be in the `bank` root directory. A custom file can be passed with the flag `--secureid`.

## Deactivate operator authentication

In order to create the first account that will play the role of operator, we need to temporarily deactivate authentication.

This can be done by modifying `bank/api/handlers/userCreate.go`, switching the `withOperatorAuth` const to `false`.

## Run the api service

`go run ./api/cmd/bank-api`

## Run the accounting service

`go run ./accounting/cmd/bankaccounting`

## Run userCreate command to create operator account

`go run ./api/cmd/bank-user-manager/ userCreate --pgpPublicKey=<some key file>`

Decrypt the message to obtain the account number and totp secret. Create a `.env` file with the following content:
```
CONDENSAT_OPERATOR_ACCOUNT=<account number>
```

Turn `withOperatorAuth` const back to `true`, and start the stack again.

You should be able to create a new account by running the `userCreate` command again and providing the totp when prompted.
