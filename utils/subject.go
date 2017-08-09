package utils

import (
	"fmt"
	"strings"
)

type Subject struct {
	preamble string
	sensor   string
	parts    []string
}

const (
	preamble = "data://"
)

func BuildSubjectKey(sensor string, key ...string) Subject {
	return Subject{
		sensor: sensor,
		parts:  key,
	}
}

func ParseSubject(subject string) (Subject, error) {

	if strings.Index(subject, preamble) != 0 {
		return Subject{}, fmt.Errorf("invalid Subject: %s", subject)
	}
	bare := strings.Trim(subject, preamble)
	bits := strings.Split(bare, "/")

	if len(bits) == 1 {
		return Subject{
			sensor: bits[0],
			parts:  []string{},
		}, nil
	}

	return Subject{
		sensor: bits[0],
		parts:  bits[1:],
	}, nil
}

func (s Subject) String() string {

	if len(s.parts) > 0 {
		return fmt.Sprintf("%s%s/%s", preamble, s.sensor, strings.Join(s.parts, "/"))
	}
	return fmt.Sprintf("%s%s", preamble, s.sensor)

}

func (s Subject) IsRoot() bool {
	return len(s.parts) == 0
}

func (s Subject) Perms() []string {

	all := []string{}
	current := fmt.Sprintf("%s%s", preamble, s.sensor)

	all = append(all, current)

	for _, s := range s.parts {

		current = fmt.Sprintf("%s/%s", current, s)
		all = append(all, current)

	}

	reverse(all)
	return all
}

func reverse(ss []string) {
	last := len(ss) - 1
	for i := 0; i < len(ss)/2; i++ {
		ss[i], ss[last-i] = ss[last-i], ss[i]
	}
}
