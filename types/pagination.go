package types

// PaginationOptions defines the options for pagination
type PaginationOptions struct {
	// Page is the page number (1-based)
	Page int `json:"page"`
	// PageSize is the number of items per page
	PageSize int `json:"pageSize"`
}

// DefaultPaginationOptions returns the default pagination options
func DefaultPaginationOptions() PaginationOptions {
	return PaginationOptions{
		Page:     1,
		PageSize: 50,
	}
}

// PaginatedResult represents a paginated list of items
type PaginatedResult struct {
	// Items is the list of items for the current page
	Items any `json:"items"`
	// TotalItems is the total number of items across all pages
	TotalItems int `json:"totalItems"`
	// TotalPages is the total number of pages
	TotalPages int `json:"totalPages"`
	// CurrentPage is the current page number
	CurrentPage int `json:"currentPage"`
	// PageSize is the number of items per page
	PageSize int `json:"pageSize"`
}

// NewPaginatedResult creates a new paginated result
func NewPaginatedResult(items any, totalItems, pageSize, currentPage int) PaginatedResult {
	totalPages := calculateTotalPages(totalItems, pageSize)

	return PaginatedResult{
		Items:       items,
		TotalItems:  totalItems,
		TotalPages:  totalPages,
		CurrentPage: currentPage,
		PageSize:    pageSize,
	}
}

// calculateTotalPages calculates the total number of pages
func calculateTotalPages(totalItems, pageSize int) int {
	if pageSize <= 0 {
		pageSize = 1
	}

	totalPages := totalItems / pageSize
	if totalItems%pageSize > 0 {
		totalPages++
	}

	return totalPages
}

// ApplyPaginationToSlice applies pagination to a slice
func ApplyPaginationToSlice(slice any, options PaginationOptions) (any, int, error) {
	// Implementation for slicing a generic slice based on pagination options
	// This is a placeholder - actual implementation would require reflection
	// to handle different slice types

	// For now, this is just a stub - real implementation would be more complex
	return slice, 0, nil
}
