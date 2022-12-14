package tasks

import (
	"context"
	"testing"

	"git.condensat.tech/bank/appcontext"
	"git.condensat.tech/bank/database"
	"git.condensat.tech/bank/messaging"
)

var testContext = context.Background()

func init() {
	dbArg := database.DefaultOptions()
	dbArg.HostName = "mariadb"
	natsArg := messaging.DefaultOptions()
	natsArg.HostName = "nats"

	ctx := testContext
	ctx = appcontext.WithMessaging(ctx, messaging.NewNats(ctx, natsArg))
	ctx = appcontext.WithDatabase(ctx, database.NewDatabase(dbArg))

	migrateDatabase(ctx)

	testContext = ctx
}

func Test_parseAssetInfo(t *testing.T) {
	if len(mockAsstInfo) == 0 {
		t.Errorf("Invalid Mock data")
	}

	type args struct {
		jsonData []byte
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// {"mockAsstInfo", args{[]byte(mockAsstInfo)}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseAssetInfo(tt.args.jsonData)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseAssetInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != 2 {
				t.Errorf("parseAssetInfo() return wrong list. %+v", got)
			}

			t.Logf("parseAssetInfo: %+v", got)
		})
	}
}

func Test_processAssetInfo(t *testing.T) {
	tests := []struct {
		name string
	}{
		// {"processAssetInfo"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processAssetInfo(testContext)
		})
	}
}

func Test_processAssetIcon(t *testing.T) {
	tests := []struct {
		name string
	}{
		// {"processAssetIcon"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processAssetIcon(testContext)
		})
	}
}

func migrateDatabase(ctx context.Context) {
	db := appcontext.Database(ctx)
	_ = db.Migrate(database.AssetModel())
}

const mockAsstInfo = `{
  "123465c803ae336c62180e52d94ee80d80828db54df9bedbb9860060f49de2eb": {
    "asset_id": "123465c803ae336c62180e52d94ee80d80828db54df9bedbb9860060f49de2eb",
    "contract": {
      "entity": {
        "domain": "scamcoinbot.com"
      },
      "issuer_pubkey": "035d0f7b0207d9cc68870abfef621692bce082084ed3ca0c1ae432dd12d889be01",
      "name": "Scamcoinbot token",
      "nonce": "57258",
      "precision": 0,
      "ticker": "SCAM",
      "version": 0
    },
    "issuance_txin": {
      "txid": "27e6bd36daef786775768a6b106053d0f2f10e03b6f278715931caa00662138d",
      "vin": 0
    },
    "issuance_prevout": {
      "txid": "fc2535f2e4fc2ef1d19b832248e3edc2c3f4c4e3ee9c2bc51777bd738a6f9582",
      "vout": 10
    },
    "name": "Scamcoinbot token",
    "ticker": "SCAM",
    "precision": 0,
    "entity": {
      "domain": "scamcoinbot.com"
    },
    "version": 0,
    "issuer_pubkey": "035d0f7b0207d9cc68870abfef621692bce082084ed3ca0c1ae432dd12d889be01"
  },
  "4d4354944366ea1e33f27c37fec97504025d6062c551208f68597d1ed40ec53e": {
    "asset_id": "4d4354944366ea1e33f27c37fec97504025d6062c551208f68597d1ed40ec53e",
    "contract": {
      "entity": {
        "domain": "magicalcryptofriends.com"
      },
      "issuer_pubkey": "02d2b29fe8ffef6acb5e75d0cd7f9c55d502bd876434b87c39ae209fc57c57f52a",
      "name": "Magical Crypto Token",
      "nonce": "13158145",
      "precision": 0,
      "ticker": "MCT",
      "version": 0
    },
    "issuance_txin": {
      "txid": "d535ded7ce07a0bb9c61d0fefff8127da3fc4833302b05e2b8a0cf9e04446af1",
      "vin": 0
    },
    "issuance_prevout": {
      "txid": "839e819d74ac98110fce63a3dab3a1075bbddcad811e0e125641989581919ab0",
      "vout": 1
    },
    "name": "Magical Crypto Token",
    "ticker": "MCT",
    "precision": 0,
    "entity": {
      "domain": "magicalcryptofriends.com"
    },
    "version": 0,
    "issuer_pubkey": "02d2b29fe8ffef6acb5e75d0cd7f9c55d502bd876434b87c39ae209fc57c57f52a"
  }
}`
