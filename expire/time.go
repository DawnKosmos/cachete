package expire

import (
	"log"
	"time"
)

type Expirator interface {
	Expire(t time.Time) bool
}

type duration struct {
	t time.Time
}

func (d duration) Expire(t time.Time) bool {
	return t.Before(d.t)
}

func In(d time.Duration) Expirator {
	return duration{t: time.Now().Add(d)}
}

var defaultExpiration = time.Duration(6 * time.Hour)

func ChangeDefaultExpiration(d time.Duration) {
	defaultExpiration = d
}

type Tag struct {
	tag string
	t   time.Time
}

func (d Tag) Expire(t time.Time) bool {
	return t.Before(d.t)
}

func (d Tag) GetValue() string {
	return d.tag
}

func WithTag(tag ...string) Expirator {
	if len(tag) == 0 {
		log.Panicln("With Tag needs an argument")
		return nil
	}
	if len(tag) == 1 {
		return Tag{tag: tag[0], t: time.Now().Add(defaultExpiration)}
	}
	return Tags{tags: tag, t: time.Now().Add(defaultExpiration)}
}

type Tags struct {
	tags []string
	t    time.Time
}

func (d Tags) Expire(t time.Time) bool {
	return t.Before(d.t)
}

func (d Tags) GetValue() []string {
	return d.tags
}
