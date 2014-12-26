package microsoft

import (
	"encoding/xml"
	"io/ioutil"
	"strings"

	"github.com/st3v/tracerr"
	"github.com/st3v/translator"
)

// The LanguageCatalog provides a slice of languages supported by
// Microsoft's Translation API.
type LanguageCatalog interface {
	Languages() ([]translator.Language, error)
}

// The LanguageProvider retrieves the names and codes of all languages
// supported by the API.
type LanguageProvider interface {
	Codes() ([]string, error)
	Names(codes []string) ([]string, error)
}

type languageCatalog struct {
	provider  LanguageProvider
	languages []translator.Language
}

type languageProvider struct {
	router     Router
	httpClient HTTPClient
}

func newLanguageCatalog(provider LanguageProvider) LanguageCatalog {
	return &languageCatalog{
		provider: provider,
	}
}

func newLanguageProvider(authenticator Authenticator) LanguageProvider {
	return &languageProvider{
		router:     newRouter(),
		httpClient: newHTTPClient(authenticator),
	}
}

func (c *languageCatalog) Languages() ([]translator.Language, error) {
	if c.languages == nil {
		codes, err := c.provider.Codes()
		if err != nil {
			return nil, tracerr.Wrap(err)
		}

		names, err := c.provider.Names(codes)
		if err != nil {
			return nil, tracerr.Wrap(err)
		}

		for i := range codes {
			c.languages = append(
				c.languages,
				translator.Language{
					Code: codes[i],
					Name: names[i],
				})
		}
	}
	return c.languages, nil
}

func (p *languageProvider) Names(codes []string) ([]string, error) {
	payload, _ := xml.Marshal(newXMLArrayOfStrings(codes))
	uri := p.router.LanguageNamesURL() + "?locale=en"

	response, err := p.httpClient.SendRequest("POST", uri, strings.NewReader(string(payload)), "text/xml")
	if err != nil {
		return nil, tracerr.Wrap(err)
	}

	body, err := ioutil.ReadAll(response.Body)
	defer response.Body.Close()
	if err != nil {
		return nil, tracerr.Wrap(err)
	}

	result := &xmlArrayOfStrings{}
	if err := xml.Unmarshal(body, &result); err != nil {
		return nil, tracerr.Wrap(err)
	}

	return result.Strings, nil
}

func (p *languageProvider) Codes() ([]string, error) {
	response, err := p.httpClient.SendRequest("GET", p.router.LanguageCodesURL(), nil, "text/plain")
	if err != nil {
		return nil, tracerr.Wrap(err)
	}

	body, err := ioutil.ReadAll(response.Body)
	defer response.Body.Close()
	if err != nil {
		return nil, tracerr.Wrap(err)
	}

	result := &xmlArrayOfStrings{}
	if err = xml.Unmarshal(body, &result); err != nil {
		return nil, tracerr.Wrap(err)
	}

	return result.Strings, nil
}
