# BankAccountManager

Command line tool to operate with Condensat Bank Accounting Api.
Communication are made through nats messaging system.
Authentification method is based on operator accountNumber and one time password (totp).

TL;DR :

```bash
  go run ./accounting/cmd/bank-account-manager fiatDeposit --userName=8868029921 --amount=200 --currency=EUR
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

### fiatDeposit

```bash
Usage of fiatDeposit:
  -amount float
        Amount to deposit on the account
  -currency string
        Currency that we intend to deposit (in ISO4217 code notation, ie. EUR)
  -label string
        Optional label
  -userName string
        User that deposits money
```
Once deposit is made the following message is displayed on screen :
`Successfully deposited <amount> <currency> for user <userName>`
