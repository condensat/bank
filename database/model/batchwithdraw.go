package model

type BatchWithdraw struct {
	BatchID    BatchID    `gorm:"unique_index:idx_batch_withdraw;index;not null"`                     // [FK] Reference to Batch table
	WithdrawID WithdrawID `gorm:"unique_index:idx_withdraw;unique_index:idx_batch_withdraw;not null"` // [FK] Reference to Withdraw table
}
