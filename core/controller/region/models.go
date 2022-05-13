package region

type CreateRegionRequest struct {
	Name          string
	DisplayName   string
	Server        string
	Certificate   string
	IngressDomain string
	HarborID      uint `json:"harborID"`
}

type UpdateRegionRequest struct {
	Name          string
	DisplayName   string
	Server        string
	Certificate   string
	IngressDomain string
	HarborID      uint `json:"harborID"`
}
