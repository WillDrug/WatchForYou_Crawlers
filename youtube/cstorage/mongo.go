package cstorage

import (
	//"./storage" not needed
	"../cinterfaces"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"	
	"log"
	"fmt"
)
type RSSStorage struct{
	Session *mgo.Session
	DB string
}

func (rst RSSStorage) Init(URI string, DB string) (error) {
	var err error
	rst.Session, err = mgo.Dial(URI)
	rst.Session.SetMode(mgo.Monotonic, true)
	rst.DB = DB
	return err
}

func (rst RSSStorage) GetSubByLink(URI string) (cinterfaces.Feed, error) {
	c := rst.Session.DB(rst.DB).C("content")
	result = FeedSubscription{}
	err := c.Find(bson.M{"url": }).One(&result)
	return result, err
}
func (RSSStorage) UpdateSubDate(sub cinterfaces.Feed) (bool, error) {
	return true, nil
}
	
// one should implement storage interface here
func some_shit() {
	session, err := mgo.Dial("127.0.0.1:27017")
	if err != nil {
		panic(err)
	}
	defer session.Close()

	// Optional. Switch the session to a monotonic behavior.
	session.SetMode(mgo.Monotonic, true)

	c := session.DB("w4u").C("content")
	//err = c.Insert(&cstorage.FeedSubscription{"youtube", "http://test", 5, 2, nil})
	//	if err != nil {	
	//	log.Fatal(err)
	//}
	result := []FeedSubscription{}
	err = c.Find(bson.M{"parser": "youtube"}).All(&result)
	if err != nil {
		log.Fatal(err)
	}
	for i:=0;i<len(result);i++ {
		fmt.Printf(result[i].Parser)
	}
}