# BankAccountManager

Command line tool to operate with Condensat Bank Accounting Api.
Communication are made through nats messaging system.
Authentification method is based on operator accountNumber and one time password (totp).

TL;DR :

```bash
  go run ./accounting/cmd/bank-account-manager fiatDeposit --userName=8868029921 --amount=200 --currency=EUR
  go run ./accounting/cmd/bank-account-manager fiatFetchPendingWithdraw
  go run ./accounting/cmd/bank-account-manager fiatFinalizeWithdraw --id 1
  go run ./accounting/cmd/bank-account-manager cryptoFetchPendingWithdraw
  go run ./accounting/cmd/bank-account-manager cryptoValidateWithdraw 1 2 3 4
  go run ./accounting/cmd/bank-account-manager cryptoCancelWithdraw --id 1
  go run ./accounting/cmd/bank-account-manager fiatCancelWithdraw --id 1
```

If `operatorAccount` and `totp` are not set, unauthenticated call is made.

## Environement variables

Use `.env` file to store operator account and nats address (tor)

```bash
  CONDENSAT_OPERATOR_ACCOUNT=123456789
  CONDENSAT_NATS_TOR=nats-host.onion
```

## Common flags

```bash
Usage of fiatDeposit:
  -natsHost string
        Nats hostName (default 'nats') (default "nats")
  -natsPort int
        Nats port (default 4222) (default 4222)
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

### fiatFetchPendingWithdraw

```
Withdraw #0: 
UserName: <userName>
IBAN: <iban>
BIC: <bic>
Currency: <currency>
Amount: <amount>
```

### fiatFinalizeWithdraw

```bash
Usage of fiatFinalizeWithdraw:
  -id uint64
      id of the withdraw obtained with fiatFetchPendingWithdraw
```

`Successfully finalized withdrawal from user <userName> to account <iban>`

### cryptoFetchPendingWithdraw

```bash
Withdraw #0
UserName: 12345678901
Address: bc1000000000000000000000000
Currency: BTC 
Amount: 100
```

### cryptoValidateWithdraw

```bash
Usage of cryptoValidateWithdraw:
  -id []uint64
      ids of the withdraws obtained with cryptoValidateWithdraw
```

#### Note
This is not a flag, just write all the space-separated ids to validate without `--id` flag.

```bash
Successfully validated withdraw #0
UserName: 12345678901
Address: bc10000000000000000000
Currency: BTC
Amount: 100
```

### cryptoCancelWithdraw

```bash
Usage of cryptoCancelWithdraw:
  -comment string
        comment about the cancel operation
  -id uint
        id of the operation we\'re canceling
```

```bash
Successfully canceled withdraw #0
AccountID: 1 
Address: bc10000000000000000000
Chain: BTC
Amount: 100
```

### fiatCancelWithdraw

```bash
Usage of cryptoCancelWithdraw:
  -comment string
        comment about the cancel operation
  -id uint
        id of the operation we're canceling
```

```bash
Successfully canceled withdraw #0
UserName: 12345678901
IBAN: FR76XXXXX
Currency: CHF
Amount: 100
```

### Notes

* For now we don't allow to register another pending withdraw for the same user and iban if another is already pending. The existing pending withdraw must first be finalized with `fiatFinalizeWithdraw` before we can use `fiatWithdraw` again. It is possible to have many withdraws pending for the same user and different beneficiaries though.
