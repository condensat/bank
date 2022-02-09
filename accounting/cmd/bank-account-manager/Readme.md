# BankAccountManager

Command line tool to operate with Condensat Bank Accounting Api.
Communication are made through nats messaging system.
Authentification method is based on operator accountNumber and one time password (totp).

TL;DR :

```bash
  go run go run ./api/cmd/bank-account-manager fiatwithdraw --pgpPublicKey=<UserPGPPublicFile> | tee -a userCreate.log
```

If `operatorAccount` and `totp` are not set, unauthenticated call is made.

## Environement variable

Use `.env` file to store operator account and nats address (tor)

```bash
  CONDENSAT_OPERATOR_ACCOUNT=123456789
  CONDENSAT_NATS_TOR=nats-host.onion
```

## Commond flags

```bash
Usage of userCreate:

  -natsHost string
    	Nats hostName (default 'nats')
  -natsPort int
    	Nats port (default 4222)

  -operatorAccount string
    	Operator Account
  -totp string
    	Operator TOTP
```

## Commands

### fiatwithdraw

```bash
Usage of fiatwithdraw:
  -pgpPublicKey string
    	Client PGP public key filename
```

Once user created pgp encrypted message is displayed with new created accountNumber.
User's public key is used for cyphering and store to database for further use.
