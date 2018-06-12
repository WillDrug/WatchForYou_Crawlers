package cstorage
type User struct {
	Stub string
}

type FeedSubscription struct {
	Parser      string
	URL         string
	LastDelta   int
	LastUpdate  int
	Subscribers []User
}
