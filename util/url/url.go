package url

import (
	netURL "net/url"
)

func MustJoinPath(base string, elem ...string) string {
	result, err := netURL.JoinPath(base, elem...)
	if err != nil {
		panic(err)
	}

	return result
}
