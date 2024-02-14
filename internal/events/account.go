package events

func init() {
	register[AccountCreated]("account.created")
	register[AccountProfileUpdated]("account.profile_updated")
	register[AccountSessionOpened]("account.session_opened")
	register[AccountEmailAdded]("account.email_added")
	register[AccountEmailVerified]("account.email_verified")
	register[AccountEmailRemoved]("account.email_removed")
}

type (
	AccountCreated struct {
		Email string `json:"email"`
		ID    int64  `json:"id"`
	}

	AccountProfileUpdated struct {
		Name     string `json:"name"`
		About    string `json:"about"`
		Callsign string `json:"callsign"`
		ID       int64  `json:"id"`
	}

	AccountSessionOpened struct {
		UserAgent string `json:"user_agent"`
		IP        string `json:"ip"`
		ID        int64  `json:"id"`
	}

	AccountEmailAdded struct {
		Email string `json:"email"`
		ID    int64  `json:"id"`
	}

	AccountEmailVerified struct {
		Email string `json:"email"`
		ID    int64  `json:"id"`
	}

	AccountEmailRemoved struct {
		Email string `json:"email"`
	}
)
