package gofeed_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"strings"
	"sync"
	"testing"

	"github.com/skuzzymiglet/gofeed"
	"github.com/stretchr/testify/assert"
)

type feedTest struct {
	file      string
	feedType  string
	feedTitle string
	hasError  bool
}

var feedTests = []feedTest{
	{"atom03_feed.xml", "atom", "Feed Title", false},
	{"atom10_feed.xml", "atom", "Feed Title", false},
	{"rss_feed.xml", "rss", "Feed Title", false},
	{"rss_feed_bom.xml", "rss", "Feed Title", false},
	{"rss_feed_leading_spaces.xml", "rss", "Feed Title", false},
	{"rdf_feed.xml", "rss", "Feed Title", false},
	{"sample.json", "json", "title", false},
	{"unknown_feed.xml", "", "", true},
	{"invalid.json", "", "", true},
}

func TestParser_Parse(t *testing.T) {
	for _, test := range feedTests {
		fmt.Printf("Testing %s... ", test.file)

		// Get feed content
		path := fmt.Sprintf("testdata/parser/universal/%s", test.file)
		f, _ := ioutil.ReadFile(path)

		// Get actual value
		fp := gofeed.NewParser()
		feed, err := fp.Parse(bytes.NewReader(f))

		if test.hasError {
			assert.NotNil(t, err)
			assert.Nil(t, feed)
		} else {
			assert.NotNil(t, feed)
			assert.Nil(t, err)
			assert.Equal(t, feed.FeedType, test.feedType)
			assert.Equal(t, feed.Title, test.feedTitle)
		}
	}
}

func TestParser_ParseString(t *testing.T) {

	for _, test := range feedTests {
		fmt.Printf("Testing %s... ", test.file)

		// Get feed content
		path := fmt.Sprintf("testdata/parser/universal/%s", test.file)
		f, _ := ioutil.ReadFile(path)

		// Get actual value
		fp := gofeed.NewParser()
		feed, err := fp.ParseString(string(f))

		if test.hasError {
			assert.NotNil(t, err)
			assert.Nil(t, feed)
		} else {
			assert.NotNil(t, feed)
			assert.Nil(t, err)
			assert.Equal(t, feed.FeedType, test.feedType)
			assert.Equal(t, feed.Title, test.feedTitle)
		}
	}
}

func ExampleParser_Parse() {
	feedData := `<rss version="2.0">
<channel>
<title>Sample Feed</title>
</channel>
</rss>`
	fp := gofeed.NewParser()
	feed, err := fp.Parse(strings.NewReader(feedData))
	if err != nil {
		panic(err)
	}
	fmt.Println(feed.Title)
}

func ExampleParser_ParseString() {
	feedData := `<rss version="2.0">
<channel>
<title>Sample Feed</title>
</channel>
</rss>`
	fp := gofeed.NewParser()
	feed, err := fp.ParseString(feedData)
	if err != nil {
		panic(err)
	}
	fmt.Println(feed.Title)
}
