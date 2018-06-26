package cinterfaces

import (
	"time"
)

// Configurator
type ConfigLoader interface {
	GetString(query string) string
	GetInt(query string) int
	GetBool(query string) bool
}

// RSS Parsing
type Entry struct {
	Title string `json:"title"`
	Description string `json:"desc"`
	URL string `json:"url"`
	Updated int64 `json:"updated"`	
}
type RSSGetter interface {
	GetRSSLinkByName(query string) (string, error)
	GetRSSLinkByLink(url string) (string, error)
	GetRSSLinkUniversal(req string) (string, error)
	CheckRSSLink(url string) (bool, error)
	GetFeedUpdates(feedURI string, since time.Time) ([]Entry, error)
	GetFeedUpdatesWithPOSIX(feedURI string, since int64) ([]Entry, error)
}

// Connectors
type Message struct {
	Content []byte
	Context string // TODO: check this bollocks
}

type Connection interface {
	Connect(cstr string) error
	Disconnect()
	Consume(qname string) error
	InboundMessages(workers int) chan Message
	ReplyToMessage(content []byte, context Message) error
}
