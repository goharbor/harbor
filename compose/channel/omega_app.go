package channel

import (
	"net/http"
)

type OmegaClient struct {
	Client    *http.Client
	Principle string
	Password  string
	Token     string
}

type OmegaAppOutput struct {
	Client *OmegaAppOutput
}

func (app *OmegaAppOutput) Create() error {
	return nil
}

func (app *OmegaAppOutput) Stop() error {
	return nil
}

func (app *OmegaAppOutput) Scale() error {
	return nil
}

func (app *OmegaAppOutput) Restart() error {
	return nil
}
