package events

type (
	AccountCreated struct {
		ID    int64
		Email string
	}

	AccountProfileUpdated struct {
		ID       int64
		Name     string
		About    string
		Callsign string
	}

	AccountSessionOpened struct {
		ID        int64
		UserAgent string
		IP        string
	}
)

func (AccountCreated) Event() string        { return "account.created" }
func (AccountProfileUpdated) Event() string { return "account.profile_updated" }
func (AccountSessionOpened) Event() string  { return "account.session_opened" }
