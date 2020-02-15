package analyzer

import (
	"bufio"
	"bytes"
	"net"
	"net/textproto"
	"regexp"
	"sort"
	"strings"
	"time"
)

var receivedFromRegexp = regexp.MustCompile(`from\s+(.*?)\s+by(.*?)(?:(?:with|via)(.*?)(?:\sid\s|$)|\sid\s|$)`)
var receivedByRegexp = regexp.MustCompile(`by(.*?)(?:(?:with|via)(.*?)(?:;|\sid\s|$)|\sid\s)`)

type Hop struct {
	From  string        `json:"from"`
	By    string        `json:"by"`
	With  string        `json:"with"`
	Time  time.Time     `json:"time"`
	Delay time.Duration `json:"delay"`
}

type Hops []Hop

func (h Hops) Len() int {
	return len(h)
}

func (h Hops) Less(i, j int) bool {
	timeA := h[i].Time
	timeB := h[j].Time
	return timeA.Equal(timeB) || timeA.Before(timeB)
}

func (h Hops) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

type Content struct {
	Headers   textproto.MIMEHeader `json:"headers"`
	Hops      []Hop                `json:"hops"`
	Source    net.IP               `json:"source"`
	From      string               `json:"from"`
	To        string               `json:"to"`
	Cc        string               `json:"cc"`
	Subject   string               `json:"subject"`
	MessageID string               `json:"message_id"`
	Date      time.Time            `json:"date"`
}

func ParseHeaders(raw []byte) (textproto.MIMEHeader, error) {
	r := textproto.NewReader(bufio.NewReader(bytes.NewReader(raw)))
	return r.ReadMIMEHeader()
}

func ParseHops(hdrs textproto.MIMEHeader) (Hops, error) {
	hops := Hops{}
	rcvd := hdrs["Received"]
	for _, v := range rcvd {
		var hop Hop
		var err error
		// Received headers should have the timestamp occur at the end of the
		// line and after a semicolon; IE. "Received: <details>; <timestamp>"
		parts := strings.Split(v, ";")
		hop.Time, err = parseDate(strings.Trim(parts[len(parts)-1], " "))
		if err != nil {
			return []Hop{}, err
		}
		if strings.HasPrefix(v, "from") {
			m := receivedFromRegexp.FindStringSubmatch(v)
			if len(m) > 0 {
				hop.From = m[1]
				hop.By = m[2]
				hop.With = m[3]
			}
		} else {
			m := receivedByRegexp.FindStringSubmatch(v)
			if len(m) > 0 {
				hop.By = m[1]
				hop.With = m[2]
			}
		}
		hops = append(hops, hop)
	}
	sort.Sort(hops)
	for i := 0; i < hops.Len(); i++ {
		if i+1 < hops.Len() {
			hops[i+1].Delay = hops[i+1].Time.Sub(hops[i].Time)
		}
	}
	return hops, nil
}

func Analyze(raw []byte) (Content, error) {
	var c Content
	var err error
	if c.Headers, err = ParseHeaders(raw); err != nil {
		return Content{}, err
	}
	if c.Hops, err = ParseHops(c.Headers); err != nil {
		return Content{}, err
	}
	date := c.Headers.Get("Date")
	if c.Date, err = parseDate(date); err != nil {
		return Content{}, err
	}
	c.Source = net.ParseIP(strings.Trim(c.Headers.Get("X-Originating-Ip"), "[]"))
	c.From = c.Headers.Get("from")
	c.To = c.Headers.Get("to")
	c.Cc = c.Headers.Get("cc")
	c.Subject = c.Headers.Get("Subject")
	c.MessageID = c.Headers.Get("Message-ID")
	return c, nil
}
