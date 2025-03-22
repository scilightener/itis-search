package pkg

import (
	"log"
	"net/url"
	"strings"
)

func PrepareLink(link string) string {
	link = strings.Trim(link, " \n\t\r")
	parsedURL, err := url.Parse(link)
	if err != nil {
		log.Fatalln("prepareLink:", link, err)
		return link
	}
	parsedURL.RawQuery = ""
	parsedURL.Fragment = ""
	return parsedURL.String()
}
