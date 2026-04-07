package dto

type RegisterServiceRequest struct {
	Audience    string         `json:"audience"     validate:"required"`
	ServiceName string         `json:"service_name" validate:"required"`
	Description string         `json:"description"`
	Permissions map[string]any `json:"permissions"`
}

type AssignServiceRequest struct {
	Audience string `json:"audience" validate:"required"`
}

type ServiceListResponse struct {
	Services []string `json:"services"`
}
