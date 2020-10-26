package gofeed

import (
	"bytes"
	"errors"
	"io"
	"strings"

	"github.com/skuzzymiglet/gofeed/atom"
	"github.com/skuzzymiglet/gofeed/json"
	"github.com/skuzzymiglet/gofeed/rss"
)

// ErrFeedTypeNotDetected is returned when the detection system can not figure
// out the Feed format
var ErrFeedTypeNotDetected = errors.New("Failed to detect feed type")

// Parser is a universal feed parser that detects
// a given feed type, parsers it, and translates it
// to the universal feed type.
type Parser struct {
	AtomTranslator Translator
	RSSTranslator  Translator
	JSONTranslator Translator
	rp             *rss.Parser
	ap             *atom.Parser
	jp             *json.Parser
}

// NewParser creates a universal feed parser.
func NewParser() *Parser {
	fp := Parser{
		rp: &rss.Parser{},
		ap: &atom.Parser{},
		jp: &json.Parser{},
	}
	return &fp
}

// Parse parses a RSS or Atom or JSON feed into
// the universal gofeed.Feed.  It takes an
// io.Reader which should return the xml/json content.
func (f *Parser) Parse(feed io.Reader) (*Feed, error) {
	// Wrap the feed io.Reader in a io.TeeReader
	// so we can capture all the bytes read by the
	// DetectFeedType function and construct a new
	// reader with those bytes intact for when we
	// attempt to parse the feeds.
	var buf bytes.Buffer
	tee := io.TeeReader(feed, &buf)
	feedType := DetectFeedType(tee)

	// Glue the read bytes from the detect function
	// back into a new reader
	r := io.MultiReader(&buf, feed)

	switch feedType {
	case FeedTypeAtom:
		return f.parseAtomFeed(r)
	case FeedTypeRSS:
		return f.parseRSSFeed(r)
	case FeedTypeJSON:
		return f.parseJSONFeed(r)
	}

	return nil, ErrFeedTypeNotDetected
}

// ParseString parses a feed XML string and into the
// universal feed type.
func (f *Parser) ParseString(feed string) (*Feed, error) {
	return f.Parse(strings.NewReader(feed))
}

func (f *Parser) parseAtomFeed(feed io.Reader) (*Feed, error) {
	af, err := f.ap.Parse(feed)
	if err != nil {
		return nil, err
	}
	return f.atomTrans().Translate(af)
}

func (f *Parser) parseRSSFeed(feed io.Reader) (*Feed, error) {
	rf, err := f.rp.Parse(feed)
	if err != nil {
		return nil, err
	}

	return f.rssTrans().Translate(rf)
}

func (f *Parser) parseJSONFeed(feed io.Reader) (*Feed, error) {
	jf, err := f.jp.Parse(feed)
	if err != nil {
		return nil, err
	}
	return f.jsonTrans().Translate(jf)
}

func (f *Parser) atomTrans() Translator {
	if f.AtomTranslator != nil {
		return f.AtomTranslator
	}
	f.AtomTranslator = &DefaultAtomTranslator{}
	return f.AtomTranslator
}

func (f *Parser) rssTrans() Translator {
	if f.RSSTranslator != nil {
		return f.RSSTranslator
	}
	f.RSSTranslator = &DefaultRSSTranslator{}
	return f.RSSTranslator
}

func (f *Parser) jsonTrans() Translator {
	if f.JSONTranslator != nil {
		return f.JSONTranslator
	}
	f.JSONTranslator = &DefaultJSONTranslator{}
	return f.JSONTranslator
}
