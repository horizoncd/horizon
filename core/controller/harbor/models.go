package harbor

import "g.hz.netease.com/horizon/pkg/harbor/models"

type Harbor struct {
	ID              uint   `json:"id"`
	Name            string `json:"name"`
	Server          string `json:"server"`
	Token           string `json:"token"`
	PreheatPolicyID int    `json:"preheatPolicyID"`
}

type Harbors []*Harbor

func ofHarborModels(entities []*models.Harbor) Harbors {
	harbors := make([]*Harbor, 0)
	for _, entity := range entities {
		harbors = append(harbors, &Harbor{
			ID:              entity.ID,
			Name:            entity.Name,
			Server:          entity.Server,
			Token:           entity.Token,
			PreheatPolicyID: entity.PreheatPolicyID,
		})
	}

	return harbors
}

type CreateHarborRequest struct {
	Name            string `json:"name"`
	Server          string `json:"server"`
	Token           string `json:"token"`
	PreheatPolicyID int    `json:"preheatPolicyID"`
}

type UpdateHarborRequest struct {
	ID              uint   `json:"id"`
	Name            string `json:"name"`
	Server          string `json:"server"`
	Token           string `json:"token"`
	PreheatPolicyID int    `json:"preheatPolicyID"`
}
