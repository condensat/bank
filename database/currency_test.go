package database

import (
	"reflect"
	"sort"
	"testing"

	"git.condensat.tech/bank"
	"git.condensat.tech/bank/database/model"
)

func Test_currencyColumnNames(t *testing.T) {
	t.Parallel()

	fields := getSortedTypeFileds(reflect.TypeOf(model.Currency{}))
	names := currencyColumnNames()
	sort.Strings(names)

	if !reflect.DeepEqual(names, fields) {
		t.Errorf("columnsNames() = %v, want %v", names, fields)
	}
}

func TestCurrency(t *testing.T) {
	const databaseName = "TestAddCurrency"
	t.Parallel()

	db := setup(databaseName, CurrencyModel())
	defer teardown(db, databaseName)

	entries := createTestData()

	// check if table is empty
	if count := CountCurrencies(db); count != 0 {
		t.Errorf("Missing CountCurrencies() = %+v, want %+v", count, 0)
	}
	defer checkFinalState(t, db, entries)

	type args struct {
		currency model.Currency
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"NotAvailable", args{entries[0]}, false},
		{"Available", args{entries[1]}, false},
		{"NotCrypto", args{entries[2]}, false},
		{"Crypto", args{entries[3]}, false},
		{"Precision0", args{entries[4]}, false},
		{"Precision12", args{entries[5]}, false},
	}
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {

			{
				// Create Tests
				got, err := AddOrUpdateCurrency(db, tt.args.currency)
				if (err != nil) != tt.wantErr {
					t.Errorf("AddOrUpdateCurrency() error = %v, wantErr %v", err, tt.wantErr)
				}
				if !reflect.DeepEqual(got, tt.args.currency) {
					t.Errorf("GetCurrency() = %+v, want %+v", got, tt.args.currency)
				}

				got, err = GetCurrencyByName(db, tt.args.currency.Name)
				if err != nil {
					t.Errorf("GetCurrencyByName() failed error = %v", err)
				}
				if !reflect.DeepEqual(got, tt.args.currency) {
					t.Errorf("GetCurrencyByName() = %+v, want %+v", got, tt.args.currency)
				}
			}

			// Exists Tests
			{
				if !CurrencyExists(db, tt.args.currency.Name) {
					t.Errorf("CurrencyExists() = %s should exists", tt.args.currency.Name)
				}
			}

			// Update Tests
			{
				updateCurr, err := GetCurrencyByName(db, tt.args.currency.Name)
				if err != nil {
					t.Errorf("GetCurrencyByName() failed error = %v", err)
				}
				// change entry
				*updateCurr.Available = 2

				got, err := AddOrUpdateCurrency(db, updateCurr)
				if (err != nil) != tt.wantErr {
					t.Errorf("AddOrUpdateCurrency() error = %v, wantErr %v", err, tt.wantErr)
				}
				if !reflect.DeepEqual(got, updateCurr) {
					t.Errorf("AddOrUpdateCurrency() = %+v, want %+v", got, updateCurr)
				}

				got, err = GetCurrencyByName(db, updateCurr.Name)
				if err != nil {
					t.Errorf("GetCurrencyByName() failed error = %v", err)
				}
				if !reflect.DeepEqual(got, updateCurr) {
					t.Errorf("GetCurrencyByName() = %+v, want %+v", got, updateCurr)
				}

				updateCurr, err = GetCurrencyByName(db, tt.args.currency.Name)
				if err != nil {
					t.Errorf("GetCurrencyByName() failed error = %v", err)
				}
				// restore entry
				*updateCurr.Available = *tt.args.currency.Available

				_, err = AddOrUpdateCurrency(db, updateCurr)
				if err != nil {
					t.Errorf("WTF")
				}
				got, err = GetCurrencyByName(db, updateCurr.Name)
				if err != nil {
					t.Errorf("GetCurrencyByName() failed error = %v", err)
				}
				if !reflect.DeepEqual(got, updateCurr) {
					t.Errorf("GetCurrencyByName() = %+v, want %+v", got, updateCurr)
				}
			}

		})
	}
}

func createTestData() []model.Currency {
	return []model.Currency{
		model.NewCurrency("USD", "", 0, 0, 1, 2),
		model.NewCurrency("BTC", "", 0, 1, 1, 2),
		model.NewCurrency("USD2", "", 2, 0, 0, 2),
		model.NewCurrency("BTC2", "", 2, 1, 1, 2),
		model.NewCurrency("USD3", "", 2, 0, 1, 0),
		model.NewCurrency("BTC3", "", 1, 1, 1, 12),
	}
}

func checkFinalState(t *testing.T, db bank.Database, entries []model.Currency) {
	// check if table has entries
	if count := CountCurrencies(db); count != len(entries) {
		t.Errorf("Missing CountCurrencies() = %+v, want %+v", count, len(entries))
	}

	{
		list, err := ListAllCurrency(db)
		if err != nil {
			t.Errorf("ListAllCurrency() Failed = %+v", err)
		}
		if len(list) != len(entries) {
			t.Errorf("Missing ListAllCurrency() = %+v, want %+v", len(list), len(entries))
		}
	}

	{
		list, err := ListAvailableCurrency(db)
		if err != nil {
			t.Errorf("ListAvailableCurrency() Failed = %+v", err)
		}
		if len(list) != len(entries)/2 {
			t.Errorf("Missing ListAvailableCurrency() = %+v, want %+v", len(list), len(entries)/2)
		}

		for _, curr := range list {
			if !curr.IsAvailable() {
				t.Errorf("Currency IsAvailable must be true")
			}
		}
	}
}
