package winterfaces


// Configurator
type ConfigLoader interface {
	SetDefaults(parms map[string]interface{}) error
	GetString(query string) string
	GetInt(query string) int
	GetBool(query string) bool
}

 
// Redundant: Now replaced by plugin interface
/*type RSSGetter interface {
	GetRSSLinkByName(query string) (string, error)
	GetRSSLinkByLink(url string) (string, error)
	GetRSSLinkUniversal(req string) (string, error)
	CheckRSSLink(url string) (bool, error)
	GetFeedUpdates(feedURI string, since time.Time) ([]Entry, error)
	GetFeedUpdatesWithPOSIX(feedURI string, since int64) ([]Entry, error)
}*/

// Data structs to be used by connector implementation (Request-Response pair)
type Entry struct {
	Title string `json:"title"`
	Description string `json:"desc"`
	URL string `json:"url"`
	Updated int64 `json:"updated"`	
}

type RPCRequest struct {
	Parser      string	`json:"parser"`
	RequestType	string	`json:"parsertype"`
	Request     string	`json:"request"`
	LastDelta   int		`json:"lastdelta"`
	LastUpdate  int64	`json:"lastupdate"`
	ReplyChannel chan RPCResponse
}


type RPCResponse struct {
	Success bool `json:"updated"`
	Message string `json:"message"`
	Entries []Entry	`json:"entries"`
}


// Connector
type Connection interface {
	Connect(cstr string) error  							// establish a connection
	Disconnect()			    							// always defer this
	Consume(qname string) (*chan RPCRequest, error) 		// Start listening to a queue or http
	//ReplyToMessage(RPCResponse) error 						//
}

// Specific host plugin interface
type Parser interface {
	CheckUpdates(request string, requestType string, since interface{}) ([]Entry, error) // Should support *time.Time and *int64 as base
}