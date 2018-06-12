package crss


func () GetFeedUpdates(feedURI) {
	fp := gofeed.NewParser()
	feed, err := fp.ParseURL("https://www.youtube.com/feeds/videos.xml?channel_id=UCBa659QWEk1AI4Tg--mrJ2A")
	if err != nil {
		
	}
	fmt.Println(feed.Items[0].Published)	
}