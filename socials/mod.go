package socials

type Social interface {
	GetName(string) (string, error)
	GetUsername(string) (string, error)
	IntoTeleEmbedLink(string) (string, error)
	IsValidURL(string) error
}

func PrefixSocialMatch() map[string]Social {
	return map[string]Social{
		"https://twitter.com/":              Twitter{},
		"https://x.com/":                    Twitter{},
		"https://www.furaffinity.net/view/": FA{},
	}
}
