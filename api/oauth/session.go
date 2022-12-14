package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"git.condensat.tech/bank"
	"git.condensat.tech/bank/accounting/client"
	"git.condensat.tech/bank/appcontext"

	"git.condensat.tech/bank/api/services"
	"git.condensat.tech/bank/api/sessions"

	"git.condensat.tech/bank/database"
	"git.condensat.tech/bank/database/model"

	"github.com/markbates/goth"
)

func getSessionCookie(r *http.Request) string {
	cookie, err := r.Cookie("sessionId")
	if err != nil {
		return ""
	}

	return cookie.Value
}

func UpdateUserSession(ctx context.Context, req *http.Request, w http.ResponseWriter, user goth.User) error {
	var userID uint64
	// Database Query
	db := appcontext.Database(ctx)
	err := db.Transaction(func(db bank.Database) error {
		u, err := database.FindUserByEmail(db, model.UserEmail(user.Email))
		if err != nil {
			return err
		}

		providerID := user.UserID

		// create user if email does not exists
		if u.ID == 0 {
			u, err = database.FindOrCreateUser(db, model.User{
				Name:  model.UserName(fmt.Sprintf("%s:%s", user.Provider, providerID)),
				Email: model.UserEmail(user.Email),
			})
			if err != nil {
				return err
			}

			// automatically create accounts for new users
			go func(userID uint64) {
				list, err := client.CurrencyList(ctx)
				if err != nil {
					return
				}

				for _, currency := range list.Currencies {
					// do not create account for disableds or not autocreate currencies
					if !currency.Available || !currency.AutoCreate {
						continue
					}

					// Create account with currency
					account, err := client.AccountCreate(ctx, userID, currency.Name)
					if err != nil {
						continue
					}

					// Enable account with normal status
					_, err = client.AccountSetStatus(ctx, account.Info.AccountID, model.AccountStatusNormal.String())
					if err != nil {
						continue
					}
				}
			}(uint64(u.ID))
		}

		// store userID for cookie creation
		userID = uint64(u.ID)

		oa, err := database.FindOrCreateOAuth(db, model.OAuth{
			Provider:   user.Provider,
			ProviderID: providerID,
			UserID:     u.ID,
		})
		if err != nil {
			return err
		}

		// store oauth data
		data, err := json.Marshal(&user)
		if err != nil {
			return err
		}

		_, err = database.CreateOrUpdateOAuthData(db, model.OAuthData{
			OAuthID: oa.ID,
			Data:    string(data),
		})
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	if userID == 0 {
		return sessions.ErrInvalidUserID
	}

	// create session & cookie
	err = services.CreateSessionWithCookie(ctx, req, w, userID)
	if err != nil {
		return err
	}

	return nil
}

func RemoveSession(ctx context.Context, req *http.Request) error {
	session, err := sessions.ContextSession(ctx)
	if err != nil {
		return err
	}

	sessionID := sessions.SessionID(getSessionCookie(req))

	return session.InvalidateSession(ctx, sessionID)
}
