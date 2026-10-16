// Package clock — «сейчас» в часовом поясе семьи; Now подменяется в тестах.
package clock

import "time"

type Clock struct {
	Location *time.Location
	Now      func() time.Time
}

func New(loc *time.Location) Clock {
	return Clock{Location: loc, Now: time.Now}
}

func (c Clock) Time() time.Time { return c.Now().In(c.Location) }
