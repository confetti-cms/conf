package config

import (
	"github.com/confetti-framework/framework/support/env"
	"net/url"
)

var Auth0 = struct {
	Domain string
	ClientId string
	Audience string
}{
	Domain: env.StringOr("AUTH0_DOMAIN", "dev-esz6mwtnerekkyi5.eu.auth0.com"),
	ClientId: env.StringOr("AUTH0_CLIENT_ID", "ExdGu6Cc8OywBaTe2exE1k7sCmYUH6Fk"),
	Audience: url.QueryEscape(env.StringOr("AUTH0_AUDIENCE", "https://confetti-cms.com")),
}
