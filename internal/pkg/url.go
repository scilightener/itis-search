package pkg

import (
	"net/url"
	"strings"
)

func PrepareLink(link string) string {
	link = strings.Trim(link, " \n\t\r")
	parsedURL, err := url.Parse(link)
	if err != nil {
		return link
	}
	parsedURL.RawQuery = ""
	parsedURL.Fragment = ""
	return parsedURL.String()
}
