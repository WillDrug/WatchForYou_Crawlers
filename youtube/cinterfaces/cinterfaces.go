package cinterfaces

type ConfigLoader interface {
	GetString(query string) string
	GetInt(query string) int
	GetBool(query string) bool
}
type MessageMeta interface {
	Ack(bool) error
}

type RSSGetter interface {
	GetRSSLinkByName(query string) (string, error)
	GetRSSLinkByLink(url string) (string, error)
	GetRSSLinkUniversal(req string) (string, error)
	CheckRSSLink(url string) (bool, error)
}