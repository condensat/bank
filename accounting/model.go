package accounting

import (
	"git.condensat.tech/bank/database"
	"git.condensat.tech/bank/database/query"
)

func Models() []database.Model {
	return query.WithdrawModel()
}
