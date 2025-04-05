package models

type Permission struct {
	ID          uint64 `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type PaginatedPermission struct {
	Data        []*Permission `json:"data"`
	Count       int64         `json:"count"`
	TotalPages  int           `json:"total_pages"`
	CurrentPage int           `json:"current_page"`
	HasNextPage bool          `json:"has_next_page"`
}
