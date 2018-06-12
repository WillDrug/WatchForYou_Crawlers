package cinterfaces

type RSSGetter interface {
	GetRSSLinkByName(query string) (string, error)
	GetRSSLinkByLink(url string) (string, error)
	GetRSSLinkUniversal(req string) (string, error)
	CheckRSSLink(url string) (bool, error)
}
type User interface {
}
type Feed interface {
}

type Storage interface {
	GetSubByLink(URI string) (Feed, error)
	UpdateSubDate(sub Feed) (bool, error)
	Init(URI string, DB string) (error)
}

