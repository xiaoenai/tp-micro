package auth

import (
	tp "github.com/henrylee2cn/teleport"
)

// AccessToken access token info
type AccessToken struct {
	Id string
}

// String returns the access token string.
func (a *AccessToken) String() string {
	// TODO

	// The default implementation
	return a.Id
}

// Apply applies for access token.
func Apply() (*AccessToken, *tp.Rerror) {
	// TODO

	// The default implementation
	return new(AccessToken), nil
}

// Verify checks access token.
func Verify(accessToken string) (*AccessToken, *tp.Rerror) {
	// TODO some business code

	// The default implementation
	return new(AccessToken), nil
}
