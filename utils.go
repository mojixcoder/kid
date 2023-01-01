package kid

import "net/url"

// getPath returns request's path.
func getPath(u *url.URL) string {
	if u.RawPath != "" {
		return u.RawPath
	}
	return u.Path
}

// resolveAddress returns the address which server will run on.
func resolveAddress(addresses []string) string {
	if len(addresses) == 0 {
		return ":2376"
	}
	return addresses[0]
}
