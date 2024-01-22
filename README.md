# Net Control

Netctl provides the Net Control application

## Development Setup

To build and develop the netctl software, you'll need:

* a working editor (you're on your own here)
* mage installed (`brew install mage`)
* install templ (`go install github.com/a-h/templ/cmd/templ@latest`)
* install sqlc (`brew install sqlc`)
* install stringer (`go install golang.org/x/tools/cmd/stringer@latest`)
* install air (`go install github.com/cosmtrek/air@latest`)

Once you build the application with `mage build`. You should be able to
configure it.

* configure the email service with creds from Postmark:
  * `netctl config set service.email.account.token 'ACCOUNT_TOKEN'`
  * `netctl config set service.email.server.token 'SERVER_TOKEN'`
