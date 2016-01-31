package feed

import (
	"errors"
	"fmt"
	"strings"

	"github.com/mmcdole/go-xpp"
)

type RSSParser struct {
	BaseParser
}

func (rp *RSSParser) ParseFeed(feed string) (rss *RSSFeed, err error) {
	rp.feedSpaces = map[string]string{}
	p := xpp.NewXMLPullParser(strings.NewReader(feed))

	_, err = p.NextTag()
	if err != nil {
		return
	}

	rp.parseNamespaces(p)

	return rp.parseRoot(p)
}

func (rp *RSSParser) parseRoot(p *xpp.XMLPullParser) (*RSSFeed, error) {
	rssErr := p.Expect(xpp.StartTag, "rss")
	rdfErr := p.Expect(xpp.StartTag, "RDF")
	if rssErr != nil && rdfErr != nil {
		return nil, fmt.Errorf("%s or %s", rssErr.Error(), rdfErr.Error())
	}

	// Items found in feed root
	var rootChannel *RSSFeed
	var rootTextInput *RSSTextInput
	var rootItems []*RSSItem
	var rootImage *RSSImage

	ver := rp.parseVersion(p)

	for {
		tok, err := p.NextTag()
		if err != nil {
			return nil, err
		}

		if tok == xpp.EndTag {
			break
		}

		if tok == xpp.StartTag {

			rp.parseNamespaces(p)

			// Skip any extensions found in the feed root.
			if rp.isExtension(p) {
				p.Skip()
				continue
			}

			if p.Name == "channel" {
				rootChannel, err = rp.parseChannel(p)
				if err != nil {
					return nil, err
				}
			} else if p.Name == "item" {
				item, err := rp.parseItem(p)
				if err != nil {
					return nil, err
				}
				if rootItems == nil {
					rootItems = []*RSSItem{}
				}
				rootItems = append(rootItems, item)
			} else if p.Name == "textinput" {
				rootTextInput, err = rp.parseTextInput(p)
				if err != nil {
					return nil, err
				}
			} else if p.Name == "image" {
				rootImage, err = rp.parseImage(p)
				if err != nil {
					return nil, err
				}
			} else {
				p.Skip()
			}
		}
	}

	rssErr = p.Expect(xpp.EndTag, "rss")
	rdfErr = p.Expect(xpp.EndTag, "RDF")
	if rssErr != nil && rdfErr != nil {
		return nil, fmt.Errorf("%s or %s", rssErr.Error(), rdfErr.Error())
	}

	if rootChannel == nil {
		return nil, errors.New("No channel element found.")
	}

	if rootTextInput != nil {
		rootChannel.TextInput = rootTextInput
	}

	if rootItems != nil {
		rootChannel.Items = append(rootChannel.Items, rootItems...)
	}

	if rootImage != nil {
		rootChannel.Image = rootImage
	}

	rootChannel.Version = ver
	return rootChannel, nil
}

func (rp *RSSParser) parseChannel(p *xpp.XMLPullParser) (rss *RSSFeed, err error) {

	if err = p.Expect(xpp.StartTag, "channel"); err != nil {
		return nil, err
	}

	rss = &RSSFeed{}
	rss.Items = []*RSSItem{}
	rss.Categories = []*RSSCategory{}
	rss.Extensions = FeedExtensions{}

	for {
		tok, err := p.NextTag()
		if err != nil {
			return nil, err
		}

		if tok == xpp.EndTag {
			break
		}

		if tok == xpp.StartTag {

			rp.parseNamespaces(p)

			if rp.isExtension(p) {
				ext, err := rp.parseExtension(rss.Extensions, p)
				if err != nil {
					return nil, err
				}
				rss.Extensions = ext
			} else if p.Name == "title" {
				result, err := p.NextText()
				if err != nil {
					return nil, err
				}
				rss.Title = result
			} else if p.Name == "description" {
				result, err := p.NextText()
				if err != nil {
					return nil, err
				}
				rss.Description = result
			} else if p.Name == "link" {
				result, err := p.NextText()
				if err != nil {
					return nil, err
				}
				rss.Link = result
			} else if p.Name == "language" {
				result, err := p.NextText()
				if err != nil {
					return nil, err
				}
				rss.Language = result
			} else if p.Name == "copyright" {
				result, err := p.NextText()
				if err != nil {
					return nil, err
				}
				rss.Copyright = result
			} else if p.Name == "managingEditor" {
				result, err := p.NextText()
				if err != nil {
					return nil, err
				}
				rss.ManagingEditor = result
			} else if p.Name == "webMaster" {
				result, err := p.NextText()
				if err != nil {
					return nil, err
				}
				rss.WebMaster = result
			} else if p.Name == "pubDate" {
				result, err := p.NextText()
				if err != nil {
					return nil, err
				}
				rss.PubDate = result
				date, err := ParseDate(result)
				if err == nil {
					rss.PubDateParsed = &date
				}
			} else if p.Name == "lastBuildDate" {
				result, err := p.NextText()
				if err != nil {
					return nil, err
				}
				rss.LastBuildDate = result
				date, err := ParseDate(result)
				if err == nil {
					rss.PubDateParsed = &date
				}
			} else if p.Name == "generator" {
				result, err := p.NextText()
				if err != nil {
					return nil, err
				}
				rss.Generator = result
			} else if p.Name == "docs" {
				result, err := p.NextText()
				if err != nil {
					return nil, err
				}
				rss.Docs = result
			} else if p.Name == "ttl" {
				result, err := p.NextText()
				if err != nil {
					return nil, err
				}
				rss.TTL = result
			} else if p.Name == "rating" {
				result, err := p.NextText()
				if err != nil {
					return nil, err
				}
				rss.Rating = result
			} else if p.Name == "item" {
				result, err := rp.parseItem(p)
				if err != nil {
					return nil, err
				}
				rss.Items = append(rss.Items, result)
			} else if p.Name == "category" {
				result, err := rp.parseCategory(p)
				if err != nil {
					return nil, err
				}
				rss.Categories = append(rss.Categories, result)
			} else {
				// Skip element as it isn't an extension and not
				// part of the spec
				p.Skip()
			}
		}
	}

	if err = p.Expect(xpp.EndTag, "channel"); err != nil {
		return nil, err
	}

	return rss, nil
}

func (rp *RSSParser) parseItem(p *xpp.XMLPullParser) (item *RSSItem, err error) {

	if err = p.Expect(xpp.StartTag, "item"); err != nil {
		return nil, err
	}

	item = &RSSItem{}
	item.Categories = []*RSSCategory{}
	item.Extensions = FeedExtensions{}

	for {
		tok, err := p.NextTag()
		if err != nil {
			return nil, err
		}

		if tok == xpp.EndTag {
			break
		}

		if tok == xpp.StartTag {

			rp.parseNamespaces(p)

			if rp.isExtension(p) {
				ext, err := rp.parseExtension(item.Extensions, p)
				if err != nil {
					return nil, err
				}
				item.Extensions = ext
			} else if p.Name == "title" {
				result, err := p.NextText()
				if err != nil {
					return nil, err
				}
				item.Title = result
			} else if p.Name == "description" {
				result, err := p.NextText()
				if err != nil {
					return nil, err
				}
				item.Description = result
			} else if p.Name == "link" {
				result, err := p.NextText()
				if err != nil {
					return nil, err
				}
				item.Link = result
			} else if p.Name == "author" {
				result, err := p.NextText()
				if err != nil {
					return nil, err
				}
				item.Author = result
			} else if p.Name == "comments" {
				result, err := p.NextText()
				if err != nil {
					return nil, err
				}
				item.Comments = result
			} else if p.Name == "pubDate" {
				result, err := p.NextText()
				if err != nil {
					return nil, err
				}
				item.PubDate = result
				date, err := ParseDate(result)
				if err == nil {
					item.PubDateParsed = &date
				}
			} else if p.Name == "source" {
				result, err := rp.parseSource(p)
				if err != nil {
					return nil, err
				}
				item.Source = result
			} else if p.Name == "enclosure" {
				result, err := rp.parseEnclosure(p)
				if err != nil {
					return nil, err
				}
				item.Enclosure = result
			} else if p.Name == "guid" {
				result, err := rp.parseGuid(p)
				if err != nil {
					return nil, err
				}
				item.Guid = result
			} else if p.Name == "category" {
				result, err := rp.parseCategory(p)
				if err != nil {
					return nil, err
				}
				item.Categories = append(item.Categories, result)
			} else {
				// Skip any elements not part of the item spec
				p.Skip()
			}
		}
	}

	if err = p.Expect(xpp.EndTag, "item"); err != nil {
		return nil, err
	}

	return item, nil
}

func (rp *RSSParser) parseSource(p *xpp.XMLPullParser) (source *RSSSource, err error) {
	if err = p.Expect(xpp.StartTag, "source"); err != nil {
		return nil, err
	}

	source = &RSSSource{}
	source.URL = p.Attribute("url")

	result, err := p.NextText()
	if err != nil {
		return source, err
	}
	source.Title = result

	if err = p.Expect(xpp.EndTag, "source"); err != nil {
		return nil, err
	}
	return source, nil
}

func (rp *RSSParser) parseEnclosure(p *xpp.XMLPullParser) (enclosure *RSSEnclosure, err error) {
	if err = p.Expect(xpp.StartTag, "enclosure"); err != nil {
		return nil, err
	}

	enclosure = &RSSEnclosure{}
	enclosure.URL = p.Attribute("url")
	enclosure.Length = p.Attribute("length")
	enclosure.Type = p.Attribute("type")

	p.NextTag()

	if err = p.Expect(xpp.EndTag, "enclosure"); err != nil {
		return nil, err
	}

	return enclosure, nil
}

func (rp *RSSParser) parseImage(p *xpp.XMLPullParser) (image *RSSImage, err error) {
	if err = p.Expect(xpp.StartTag, "image"); err != nil {
		return nil, err
	}

	image = &RSSImage{}

	for {
		tok, err := p.NextTag()
		if err != nil {
			return image, err
		}

		if tok == xpp.EndTag {
			break
		}

		if tok == xpp.StartTag {
			if p.Name == "url" {
				result, err := p.NextText()
				if err != nil {
					return nil, err
				}
				image.URL = result
			} else if p.Name == "title" {
				result, err := p.NextText()
				if err != nil {
					return nil, err
				}
				image.Title = result
			} else if p.Name == "link" {
				result, err := p.NextText()
				if err != nil {
					return nil, err
				}
				image.Link = result
			} else if p.Name == "width" {
				result, err := p.NextText()
				if err != nil {
					return nil, err
				}
				image.Width = result
			} else if p.Name == "height" {
				result, err := p.NextText()
				if err != nil {
					return nil, err
				}
				image.Height = result
			} else {
				p.Skip()
			}
		}
	}

	if err = p.Expect(xpp.EndTag, "image"); err != nil {
		return nil, err
	}

	return image, nil
}

func (rp *RSSParser) parseGuid(p *xpp.XMLPullParser) (guid *RSSGuid, err error) {
	if err = p.Expect(xpp.StartTag, "guid"); err != nil {
		return nil, err
	}

	guid = &RSSGuid{}
	guid.IsPermalink = p.Attribute("isPermalink")

	result, err := p.NextText()
	if err != nil {
		return
	}
	guid.Value = result

	if err = p.Expect(xpp.EndTag, "guid"); err != nil {
		return nil, err
	}

	return guid, nil
}

func (rp *RSSParser) parseCategory(p *xpp.XMLPullParser) (cat *RSSCategory, err error) {

	if err = p.Expect(xpp.StartTag, "category"); err != nil {
		return nil, err
	}

	cat = &RSSCategory{}
	cat.Domain = p.Attribute("domain")

	result, err := p.NextText()
	if err != nil {
		return nil, err
	}

	cat.Value = result

	if err = p.Expect(xpp.EndTag, "category"); err != nil {
		return nil, err
	}
	return cat, nil
}

func (rp *RSSParser) parseTextInput(p *xpp.XMLPullParser) (ti *RSSTextInput, err error) {
	if err = p.Expect(xpp.StartTag, "textinput"); err != nil {
		return nil, err
	}

	ti = &RSSTextInput{}

	for {
		tok, err := p.NextTag()
		if err != nil {
			return nil, err
		}

		if tok == xpp.EndTag {
			break
		}

		if tok == xpp.StartTag {
			if p.Name == "title" {
				result, err := p.NextText()
				if err != nil {
					return nil, err
				}
				ti.Title = result
			} else if p.Name == "description" {
				result, err := p.NextText()
				if err != nil {
					return nil, err
				}
				ti.Description = result
			} else if p.Name == "name" {
				result, err := p.NextText()
				if err != nil {
					return nil, err
				}
				ti.Name = result
			} else if p.Name == "link" {
				result, err := p.NextText()
				if err != nil {
					return nil, err
				}
				ti.Link = result
			} else {
				p.Skip()
			}
		}
	}

	if err = p.Expect(xpp.EndTag, "image"); err != nil {
		return nil, err
	}

	return ti, nil
}

func (rp *RSSParser) parseVersion(p *xpp.XMLPullParser) (ver string) {
	name := strings.ToLower(p.Name)
	if name == "rss" {
		ver = p.Attribute("version")
		if ver == "" {
			ver = "2.0"
		}
	} else if name == "rdf" {
		ns := p.Attribute("xmlns")
		if ns == "http://channel.netscape.com/rdf/simple/0.9/" {
			ver = "0.9"
		} else if ns == "http://purl.org/rss/1.0/" {
			ver = "1.0"
		}
	}
	return
}