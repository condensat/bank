package database

import (
	"reflect"
	"testing"
	"time"

	"git.condensat.tech/bank/database/model"
)

func TestAddWithdrawValidation(t *testing.T) {
	const databaseName = "TestAddWithdrawValidation"
	t.Parallel()

	db := setup(databaseName, WithdrawModel())
	defer teardown(db, databaseName)

	data := createTestAccountStateData(db)
	a1 := data.Accounts[0]
	a2 := data.Accounts[2]

	w1, _ := AddWithdraw(db, a1.ID, a2.ID, 10.15, model.BatchModeNormal, "{}")
	w2, _ := AddWithdraw(db, a2.ID, a1.ID, 100.15, model.BatchModeNormal, "{}")
	_, _ = AddWithdrawTarget(db, w1.ID, model.WithdrawTargetOnChain, model.WithdrawTargetData(""))
	_, _ = AddWithdrawTarget(db, w2.ID, model.WithdrawTargetSepa, model.WithdrawTargetData(""))

	amount := 10.0

	type args struct {
		userID     model.UserID
		withdrawID model.WithdrawID
		base       model.CurrencyName
		amount     model.Float
	}
	tests := []struct {
		name    string
		args    args
		want    model.Validation
		wantErr bool
	}{
		{"Defaults", args{}, model.Validation{}, true},
		{"Invalid UserID", args{userID: 0}, model.Validation{}, true},
		{"Invalid WithdrawID", args{userID: a1.UserID, withdrawID: 0}, model.Validation{}, true},
		{"Invalid base", args{userID: a1.UserID, withdrawID: w1.ID, base: ""}, model.Validation{}, true},
		{"Invalid amount", args{userID: a1.UserID, withdrawID: w1.ID, base: "CHF", amount: -1}, model.Validation{}, true},

		{"Valid Crypto", args{
			userID:     a1.UserID,
			withdrawID: w1.ID,
			base:       "CHF",
			amount:     model.Float(amount)},
			createValidation(
				a1.UserID,
				model.OperationTypeWithdraw,
				model.RefID(w1.ID),
				model.WithdrawTargetOnChain,
				"CHF",
				model.Float(amount),
			), false},
		{"Valid Fiat", args{
			userID:     a2.UserID,
			withdrawID: w2.ID,
			base:       "CHF",
			amount:     model.Float(amount)},
			createValidation(
				a2.UserID,
				model.OperationTypeWithdraw,
				model.RefID(w2.ID),
				model.WithdrawTargetSepa,
				"CHF",
				model.Float(amount),
			), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := AddWithdrawValidation(db, tt.args.userID, tt.args.withdrawID, tt.args.base, tt.args.amount)
			if (err != nil) != tt.wantErr {
				t.Errorf("AddWithdrawValidation() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if got.Timestamp.IsZero() || got.Timestamp.After(time.Now()) {
					t.Errorf("AddWithdrawValidation() wrong Timestamp %v", got.Timestamp)
				}
			}

			tt.want.ID = got.ID
			tt.want.Timestamp = got.Timestamp
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AddWithdrawValidation() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetWithdrawValidationsFromStartToNow(t *testing.T) {
	const databaseName = "TestGetWithdrawValidationsFromStartToNow"
	t.Parallel()

	db := setup(databaseName, WithdrawModel())
	defer teardown(db, databaseName)

	data := createTestAccountStateData(db)
	a1 := data.Accounts[0]
	a2 := data.Accounts[2]

	w1, _ := AddWithdraw(db, a1.ID, a2.ID, 0.1, model.BatchModeNormal, "{}")
	w2, _ := AddWithdraw(db, a2.ID, a1.ID, 0.1, model.BatchModeNormal, "{}")
	_, _ = AddWithdrawTarget(db, w1.ID, model.WithdrawTargetOnChain, model.WithdrawTargetData(""))
	_, _ = AddWithdrawTarget(db, w2.ID, model.WithdrawTargetSepa, model.WithdrawTargetData(""))

	_, _ = AddWithdrawValidation(db, a1.UserID, w1.ID, "CHF", *w1.Amount)
	_, _ = AddWithdrawValidation(db, a2.UserID, w2.ID, "CHF", *w2.Amount)

	amt := model.Float(*w1.Amount)

	type args struct {
		userID model.UserID
		start  time.Time
		target model.WithdrawTargetType
	}
	tests := []struct {
		name    string
		args    args
		want    []model.Validation
		wantErr bool
	}{
		{"Defaults", args{}, nil, true},
		{"Invalid UserID", args{userID: 0}, nil, true},
		{"Invalid start", args{userID: a1.UserID, start: time.Now().Add(time.Hour)}, nil, true},
		{"Valid", args{userID: a1.UserID, start: time.Now().Add(-24 * time.Hour)}, []model.Validation{
			{
				UserID:        a1.UserID,
				OperationType: model.OperationTypeWithdraw,
				ReferenceID:   model.RefID(w1.ID),
				Type:          model.WithdrawTargetOnChain,
				Amount:        &amt,
			},
		}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetWithdrawValidationsFromStartToNow(db, tt.args.userID, tt.args.start, tt.args.target)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetWithdrawValidationsFromStartToNow() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(got) != 0 {
				if got[0].Timestamp.IsZero() || got[0].Timestamp.After(time.Now()) {
					t.Errorf("GetWithdrawValidationsFromStartToNow() wrong Timestamp %v", got[0].Timestamp)
				}

				if *tt.want[0].Amount != *got[0].Amount {
					t.Errorf("GetWithdrawValidationsFromStartToNow() = amount: %v, want %v", *got[0].Amount, *tt.want[0].Amount)
					t.Errorf("%+v", got)
				}

				if tt.want[0].UserID != got[0].UserID {
					t.Errorf("GetWithdrawValidationsFromStartToNow() = UserID: %v, want %v", got[0].UserID, tt.want[0].UserID)
				}

				if tt.want[0].OperationType != got[0].OperationType {
					t.Errorf("GetWithdrawValidationsFromStartToNow() = OperationType: %v, want %v", got[0].OperationType, tt.want[0].OperationType)
				}
				if tt.want[0].ReferenceID != got[0].ReferenceID {
					t.Errorf("GetWithdrawValidationsFromStartToNow() = ReferenceID: %v, want %v", got[0].ReferenceID, tt.want[0].ReferenceID)
				}
				if tt.want[0].Type != got[0].Type {
					t.Errorf("GetWithdrawValidationsFromStartToNow() = Type: %v, want %v", got[0].Type, tt.want[0].Type)
				}
			}
		})
	}
}

func createValidation(userID model.UserID, operation model.OperationType, ref model.RefID, withdrawType model.WithdrawTargetType, base model.CurrencyName, amount model.Float) model.Validation {
	return model.Validation{
		UserID:        userID,
		OperationType: operation,
		ReferenceID:   ref,
		Type:          withdrawType,
		Base:          base,
		Amount:        &amount,
	}
}
