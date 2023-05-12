package models

import "time"

type Metatag struct {
	TagKey           string    `json:"tagKey"`
	TagValue         string    `json:"tagValue"`
	TagValueIdentity string    `json:"tagValueIdentity"`
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt"`
}
