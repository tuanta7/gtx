package token

import "time"

type PATStrategy struct {
	provider      string
	tokensPageURL string
}

func NewPATStrategy(provider, tokensPageURL string) *PATStrategy {
	return &PATStrategy{
		provider:      provider,
		tokensPageURL: tokensPageURL,
	}
}

func (p *PATStrategy) Provider() string {
	return PATProvider
}

func (p *PATStrategy) SaveToken(token string) error {
	return saveToken(p.provider, token)
}

func (p *PATStrategy) AuthorizeDevice() (*DeviceCodeResponse, error) {
	return nil, nil
}

func (p *PATStrategy) PollAccessToken(deviceCode string, interval time.Duration) (string, error) {
	return "", nil
}

func (p *PATStrategy) FetchUser(token string) (*User, error) {
	return nil, nil
}
