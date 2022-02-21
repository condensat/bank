package database

import (
	"reflect"
	"testing"
	"time"

	"git.condensat.tech/bank"
	"git.condensat.tech/bank/database/model"
)

func TestAddFiatOperationInfo(t *testing.T) {
	const databaseName = "TestAddFiatOperationInfo"
	t.Parallel()

	db := setup(databaseName, []model.Model{new(model.FiatOperationInfo)})
	defer teardown(db, databaseName)

	timestamp := time.Now().UTC().Truncate(time.Second)

	var amount model.Float = 25.0

	ref := model.FiatOperationInfo{
		SepaInfoID:   1,
		UserID:       1,
		CurrencyName: "CHF",
		Amount:       model.ZeroFloat(&amount),
		Type:         model.OperationTypeFiatWithdraw,
		Status:       model.FiatOperationStatusPending,
	}

	type args struct {
		db        bank.Database
		operation model.FiatOperationInfo
	}
	tests := []struct {
		name    string
		args    args
		want    model.FiatOperationInfo
		wantErr bool
	}{
		{"Default", args{}, model.FiatOperationInfo{}, true},
		{"Update", args{db, model.FiatOperationInfo{ID: 1}}, model.FiatOperationInfo{}, true},
		{"Invalid Type", args{db, model.FiatOperationInfo{Type: model.OperationTypeDeposit}}, model.FiatOperationInfo{}, true},
		{"Valid", args{db, ref}, model.FiatOperationInfo{
			ID:                1,
			SepaInfoID:        ref.SepaInfoID,
			UserID:            ref.UserID,
			CurrencyName:      ref.CurrencyName,
			Amount:            model.ZeroFloat(&amount),
			CreationTimestamp: timestamp,
			Type:              ref.Type,
			Status:            ref.Status,
		}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := AddFiatOperationInfo(db, tt.args.operation)
			if (err != nil) != tt.wantErr {
				t.Errorf("AddFiatOperationInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AddFiatOperationInfo() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFiatOperationFinalize(t *testing.T) {
	const databaseName = "TestFiatOperationFinalize"
	t.Parallel()

	db := setup(databaseName, []model.Model{new(model.FiatOperationInfo)})
	defer teardown(db, databaseName)

	entries := testFiatOperationFinalizeData()

	for _, entry := range entries {
		_, err := AddFiatOperationInfo(db, entry)
		if err != nil {
			panic(err)
		}
	}

	validReturn, err := FindFiatOperationById(db, model.FiatOperationInfoID(2))
	if err != nil {
		panic(err)
	}

	validReturn.Status = model.FiatOperationStatusComplete
	validReturn.UpdateTimestamp = time.Now().UTC().Truncate(time.Second)

	type args struct {
		operationID model.FiatOperationInfoID
	}
	tests := []struct {
		name    string
		args    args
		want    model.FiatOperationInfo
		wantErr bool
	}{
		{"Status Aborted", args{model.FiatOperationInfoID(1)}, model.FiatOperationInfo{}, true},
		{"Invalid Operation Info", args{model.FiatOperationInfoID(0)}, model.FiatOperationInfo{}, true},
		{"Valid", args{model.FiatOperationInfoID(2)}, validReturn, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FiatOperationFinalize(db, tt.args.operationID)
			if (err != nil) != tt.wantErr {
				t.Errorf("FiatOperationFinalize() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FiatOperationFinalize() = %v, want %v", got, tt.want)
			}
		})
	}
}

func testFiatOperationFinalizeData() []model.FiatOperationInfo {
	var amount model.Float = 25.0

	return []model.FiatOperationInfo{
		{
			SepaInfoID:   1,
			UserID:       1,
			CurrencyName: "CHF",
			Amount:       model.ZeroFloat(&amount),
			Type:         model.OperationTypeFiatWithdraw,
			Status:       model.FiatOperationStatusCanceled,
		},
		{
			SepaInfoID:   2,
			UserID:       2,
			CurrencyName: "CHF",
			Amount:       model.ZeroFloat(&amount),
			Type:         model.OperationTypeFiatWithdraw,
			Status:       model.FiatOperationStatusPending,
		},
	}
}

// func TestFindFiatOperationInfoByUserIDAndSepa(t *testing.T) {
// 	type args struct {
// 		db     bank.Database
// 		userID model.UserID
// 		sepaID model.SepaInfoID
// 	}
// 	tests := []struct {
// 		name    string
// 		args    args
// 		want    []model.FiatOperationInfo
// 		wantErr bool
// 	}{
// 		// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			got, err := FindFiatOperationInfoByUserIDAndSepa(tt.args.db, tt.args.userID, tt.args.sepaID)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("FindFiatOperationInfoByUserIDAndSepa() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}
// 			if !reflect.DeepEqual(got, tt.want) {
// 				t.Errorf("FindFiatOperationInfoByUserIDAndSepa() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }

// func TestFindFiatOperationPendingForUserAndSepa(t *testing.T) {
// 	type args struct {
// 		db     bank.Database
// 		userID model.UserID
// 		sepaID model.SepaInfoID
// 	}
// 	tests := []struct {
// 		name    string
// 		args    args
// 		want    []model.FiatOperationInfo
// 		wantErr bool
// 	}{
// 		// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			got, err := FindFiatOperationPendingForUserAndSepa(tt.args.db, tt.args.userID, tt.args.sepaID)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("FindFiatOperationPendingForUserAndSepa() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}
// 			if !reflect.DeepEqual(got, tt.want) {
// 				t.Errorf("FindFiatOperationPendingForUserAndSepa() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }

// func TestFindFiatWithdrawalPendingForUserAndSepa(t *testing.T) {
// 	type args struct {
// 		db     bank.Database
// 		userID model.UserID
// 		sepaID model.SepaInfoID
// 	}
// 	tests := []struct {
// 		name    string
// 		args    args
// 		want    []model.FiatOperationInfo
// 		wantErr bool
// 	}{
// 		// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			got, err := FindFiatWithdrawalPendingForUserAndSepa(tt.args.db, tt.args.userID, tt.args.sepaID)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("FindFiatWithdrawalPendingForUserAndSepa() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}
// 			if !reflect.DeepEqual(got, tt.want) {
// 				t.Errorf("FindFiatWithdrawalPendingForUserAndSepa() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }

// func TestFetchFiatPendingWithdraw(t *testing.T) {
// 	type args struct {
// 		db bank.Database
// 	}
// 	tests := []struct {
// 		name    string
// 		args    args
// 		want    []model.FiatOperationInfo
// 		wantErr bool
// 	}{
// 		// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			got, err := FetchFiatPendingWithdraw(tt.args.db)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("FetchFiatPendingWithdraw() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}
// 			if !reflect.DeepEqual(got, tt.want) {
// 				t.Errorf("FetchFiatPendingWithdraw() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }

// func TestQueryFiatOperationList(t *testing.T) {
// 	type args struct {
// 		db            bank.Database
// 		userID        model.UserID
// 		sepaID        model.SepaInfoID
// 		operationType model.OperationType
// 		status        model.FiatOperationStatus
// 	}
// 	tests := []struct {
// 		name    string
// 		args    args
// 		want    []model.FiatOperationInfo
// 		wantErr bool
// 	}{
// 		// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			got, err := QueryFiatOperationList(tt.args.db, tt.args.userID, tt.args.sepaID, tt.args.operationType, tt.args.status)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("QueryFiatOperationList() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}
// 			if !reflect.DeepEqual(got, tt.want) {
// 				t.Errorf("QueryFiatOperationList() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }

// func Test_convertFiatOperationList(t *testing.T) {
// 	type args struct {
// 		list []*model.FiatOperationInfo
// 	}
// 	tests := []struct {
// 		name string
// 		args args
// 		want []model.FiatOperationInfo
// 	}{
// 		// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			if got := convertFiatOperationList(tt.args.list); !reflect.DeepEqual(got, tt.want) {
// 				t.Errorf("convertFiatOperationList() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }

// func TestScopeFiatOperationSepaInfoID(t *testing.T) {
// 	type args struct {
// 		sepaInfoID model.SepaInfoID
// 	}
// 	tests := []struct {
// 		name string
// 		args args
// 		want func(db *gorm.DB) *gorm.DB
// 	}{
// 		// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			if got := ScopeFiatOperationSepaInfoID(tt.args.sepaInfoID); !reflect.DeepEqual(got, tt.want) {
// 				t.Errorf("ScopeFiatOperationSepaInfoID() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }

// func TestScopeFiatOperationStatus(t *testing.T) {
// 	type args struct {
// 		status model.FiatOperationStatus
// 	}
// 	tests := []struct {
// 		name string
// 		args args
// 		want func(db *gorm.DB) *gorm.DB
// 	}{
// 		// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			if got := ScopeFiatOperationStatus(tt.args.status); !reflect.DeepEqual(got, tt.want) {
// 				t.Errorf("ScopeFiatOperationStatus() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }

// func TestScopeFiatOperationType(t *testing.T) {
// 	type args struct {
// 		operationType model.OperationType
// 	}
// 	tests := []struct {
// 		name string
// 		args args
// 		want func(db *gorm.DB) *gorm.DB
// 	}{
// 		// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			if got := ScopeFiatOperationType(tt.args.operationType); !reflect.DeepEqual(got, tt.want) {
// 				t.Errorf("ScopeFiatOperationType() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }

// func Test_fiatOperationColumnNames(t *testing.T) {
// 	tests := []struct {
// 		name string
// 		want []string
// 	}{
// 		// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			if got := fiatOperationColumnNames(); !reflect.DeepEqual(got, tt.want) {
// 				t.Errorf("fiatOperationColumnNames() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }

// func Test_reqFiatOperationSepaInfoID(t *testing.T) {
// 	tests := []struct {
// 		name string
// 		want string
// 	}{
// 		// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			if got := reqFiatOperationSepaInfoID(); got != tt.want {
// 				t.Errorf("reqFiatOperationSepaInfoID() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }

// func Test_reqFiatOperationStatus(t *testing.T) {
// 	tests := []struct {
// 		name string
// 		want string
// 	}{
// 		// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			if got := reqFiatOperationStatus(); got != tt.want {
// 				t.Errorf("reqFiatOperationStatus() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }

// func Test_reqFiatOperationType(t *testing.T) {
// 	tests := []struct {
// 		name string
// 		want string
// 	}{
// 		// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			if got := reqFiatOperationType(); got != tt.want {
// 				t.Errorf("reqFiatOperationType() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }
