package dto

import "strings"

type UpdateProfileRequest struct {
	Name string `json:"name"`
}

func (r UpdateProfileRequest) Validate() map[string]string {
	fields := map[string]string{}
	name := strings.TrimSpace(r.Name)
	if len(name) < 2 || len(name) > 120 {
		fields["name"] = "Name must be between 2 and 120 characters"
	}
	return fields
}

type AdminStatusRequest struct {
	IsActive bool `json:"isActive"`
}

type AdminRoleRequest struct {
	Role string `json:"role"`
}
