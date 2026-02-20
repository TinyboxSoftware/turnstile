package config

type RouteKey string

const (
	RouteLogin    RouteKey = "login"
	RouteLogout   RouteKey = "logout"
	RouteCallback RouteKey = "callback"
	RouteHealth   RouteKey = "health"
)

type URLType bool

const (
	PathOnly URLType = false
	FullURL  URLType = true
)

var routePaths = map[RouteKey]string{
	RouteLogin:    "/oauth/login",
	RouteLogout:   "/oauth/logout",
	RouteCallback: "/oauth/callback",
	RouteHealth:   "/health",
}

func (c *Config) URI(key RouteKey, urlType URLType) string {
	path := c.AuthPrefix + routePaths[key]
	if urlType == FullURL {
		return c.PublicURL + path
	}
	return path
}
