// Package types defines common data structures used across the application.
package types

// QueryParams represents query parameters for filtering, pagination, and sorting.
type QueryParams struct {
	Search   string                 `form:"search"`
	Filter   map[string]interface{} `form:"filter"`
	Sort     string                 `form:"sort"`
	Page     int                    `form:"page"`
	PageSize int                    `form:"page_size"`
	Fields   string                 `form:"fields"`
	Include  string                 `form:"include"`
}

// Response represents a standard API response structure.
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Meta    *Meta       `json:"meta,omitempty"`
}

// Meta represents pagination metadata.
type Meta struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
	Total    int `json:"total"`
	Pages    int `json:"pages"`
}

// NewSuccessResponse creates a new success response.
func NewSuccessResponse(data interface{}, meta *Meta) Response {
	return Response{
		Success: true,
		Data:    data,
		Meta:    meta,
	}
}

// NewErrorResponse creates a new error response.
func NewErrorResponse(message string) Response {
	return Response{
		Success: false,
		Error:   message,
	}
}

// SetDefaultPagination sets default values for pagination parameters.
func (q *QueryParams) SetDefaultPagination() {
	if q.Page <= 0 {
		q.Page = 1
	}
	if q.PageSize <= 0 {
		q.PageSize = 10
	} else if q.PageSize > 100 {
		q.PageSize = 100
	}
}

// CalculatePages calculates total pages based on total records and page size.
func (m *Meta) CalculatePages() {
	if m.PageSize <= 0 {
		m.PageSize = 10
	}
	if m.Total > 0 {
		m.Pages = (m.Total + m.PageSize - 1) / m.PageSize
	} else {
		m.Pages = 0
	}
}
