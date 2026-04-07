package auth

import "github.com/jxdones/ferret/internal/collection"

// Resolve resolves the authentication configuration for a request.
func Resolve(req *collection.Auth, def *collection.Auth) *collection.Auth {
	// If the request authentication is nil or the type is empty or "inherit", return the default authentication.
	if req == nil || req.Type == "" || req.Type == "inherit" {
		return def
	}
	if req.Type == "none" {
		return nil
	}
	return req
}
