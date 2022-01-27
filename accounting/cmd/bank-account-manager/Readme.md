# BankAccountManager

Command line tool to operate with Condensat Bank Accounting Api.
Communication are made through nats messaging system.
Authentification method is based on operator accountNumber and one time password (totp).

TL;DR :

```bash
  go run ./accounting/cmd/bank-account-manager fiatDeposit --userName=8868029921 --amount=200 --currency=EUR
  go run ./accounting/cmd/bank-account-manager fiatWithdraw --userName=8868029921 --amount=20 --currency=EUR --withdrawLabel=label --iban="FR76 TEST" --bic=TEST_BIC --sepaLabel=label
  go run ./accounting/cmd/bank-account-manager fiatFetchPendingWithdraw
  go run ./accounting/cmd/bank-account-manager fiatFinalizeWithdraw --userName=8868029921 --iban="FR76 TEST"

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

### fiatWithdraw

```bash
Usage of fiatWithdraw:
  -amount float
        Amount to withdraw from the account
  -bic string
        BIC of the recipient account
  -currency string
        Currency that we intend to withdraw
  -iban string
        IBAN of the recipient account
  -sepaLabel string
        Optional Label given by the user
  -userName string
        User that ask to withdraw money
  -withdrawLabel string
        Optional Label given by the bank
```

`Successfully withdrew <amount> <currency> for user <userName>
Destination is <iban>`

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
  -iban string
        IBAN of the recipient account
  -userName string
        User that ask to withdraw money
```

`Successfully finalized withdrawal from user <userName> to account <iban>`

### Notes

* For now we don't allow to register another pending withdraw for the same user and iban if another is already pending. The existing pending withdraw must first be finalized with `fiatFinalizeWithdraw` before we can use `fiatWithdraw` again. It is possible to have many withdraws pending for the same user and different beneficiaries though.