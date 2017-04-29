package pac

import (
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
)

// SmartLoader attempt to detect if we are using js, a url, or a file path
func SmartLoader(thing string) Loader {
	return func() (string, error) {
		if strings.Contains(thing, "FindProxyForURL") && strings.Contains(thing, "{") {
			log.Print("loading pac as string")
			return thing, nil
		}
		if parseURL, parseErr := url.Parse(thing); parseErr == nil {
			switch strings.ToLower(parseURL.Scheme) {
			case "http", "https":
				return HTTPLoader(parseURL)()
			}
		}
		return FileLoader(thing)()
	}
}

func FileLoader(file string) Loader {
	return func() (string, error) {
		log.Printf("loading pac from file %q", file)
		buf, err := ioutil.ReadFile(file)
		if err != nil {
			return "", err
		}
		return string(buf), nil
	}
}

func HTTPLoader(u *url.URL) Loader {
	return func() (string, error) {
		log.Printf("loading pac from URL %q", u)
		res, err := http.Get(u.String())
		if err != nil {
			return "", err
		}
		defer res.Body.Close()
		pac, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return "", err
		}
		return string(pac), nil
	}
}
