package utils

import "time"

const (
	// ISO8601 is our date format of choice
	ISO8601 = "2006-01-02T15:04:05.999Z07:00"
)

// ISO8601Time ensures a time.Time is serialised to the ISO8601 format
type ISO8601Time struct {
	time.Time
}

func (t ISO8601Time) MarshalJSON() ([]byte, error) {
	b := make([]byte, 0, len(ISO8601)+2)
	b = append(b, '"')
	b = time.Time(t.Time).AppendFormat(b, ISO8601)
	b = append(b, '"')
	return b, nil
}

func (t *ISO8601Time) UnmarshalJSON(data []byte) (err error) {
	tt, err := time.Parse(`"`+ISO8601+`"`, string(data))

	*t = ISO8601Time{tt}
	return
}

func (t *ISO8601Time) AsTime() time.Time {

	if t == nil {
		return time.Time{}
	}
	return t.Time
}

func (t ISO8601Time) String() string {
	return t.Format(ISO8601)
}
