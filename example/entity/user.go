package entity

import "time"

type Authority string

const (
	AuthorityAdmin Authority = "admin"
	AuthorityUser  Authority = "user"
)

type Group struct {
	ID        string
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type User struct {
	ID        string
	Name      string
	Authority Authority
	Groups    []*Group `grepo:"optional:true"`
	CreatedAt time.Time
	UpdatedAt time.Time
}
