package analyzer

import (
	"time"
)

// datetimeFmts is a list of datetime formats that have been standardized or has
// been seen in an email's header.
var datetimeFmts = []string{
	time.UnixDate,
	time.RubyDate,
	time.RFC822,
	time.RFC822Z,
	time.RFC1123,
	time.RFC1123Z,
	time.RFC3339,
	time.RFC3339Nano,
	"Mon, 02 Jan 2006 15:04:05 -0700 (MST)",
	"January 02, 2006 3:04:05 PM MST",
}

// parseDate returns a time object representative of the given datetime string.
// If the format of the datetime string is unknown, an error is returned.
func parseDate(datetime string) (t time.Time, err error) {
	for _, f := range datetimeFmts {
		if t, err = time.Parse(f, datetime); err == nil {
			break
		}
	}
	return
}
