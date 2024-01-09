package badge

import (
	"time"

	"github.com/horizoncd/horizon/pkg/badge/models"
)

type Badge struct {
	Create       `json:",inline"`
	ID           uint      `json:"id"`
	ResourceID   uint      `json:"resourceID"`
	ResourceType string    `json:"resourceType"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

func (b *Badge) FromDAO(daoBadge *models.Badge) {
	b.ID = daoBadge.ID
	b.ResourceID = daoBadge.ResourceID
	b.ResourceType = daoBadge.ResourceType
	b.Name = daoBadge.Name
	b.SvgLink = daoBadge.SvgLink
	b.RedirectLink = daoBadge.RedirectLink
	b.CreatedAt = daoBadge.CreatedAt
	b.UpdatedAt = daoBadge.UpdatedAt
}

type Update struct {
	SvgLink      *string `json:"svgLink"`
	RedirectLink *string `json:"redirectLink"`
}

type Create struct {
	SvgLink      string `json:"svgLink"`
	RedirectLink string `json:"redirectLink,omitempty"`
	Name         string `json:"name"`
}
