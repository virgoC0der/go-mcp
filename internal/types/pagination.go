package types

import (
	"fmt"
	"reflect"
)

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
	sliceVal := reflect.ValueOf(slice)

	// Check if the input is a slice
	if sliceVal.Kind() != reflect.Slice {
		return nil, 0, fmt.Errorf("input is not a slice")
	}

	// Get total items count
	totalItems := sliceVal.Len()

	// Handle empty input
	if totalItems == 0 {
		// Return empty slice of the same type
		return slice, 0, nil
	}

	// Normalize page and pageSize
	if options.Page < 1 {
		options.Page = 1
	}
	if options.PageSize < 1 {
		options.PageSize = DefaultPaginationOptions().PageSize
	}

	// Calculate start and end indices
	startIdx := (options.Page - 1) * options.PageSize
	endIdx := startIdx + options.PageSize

	// Adjust if out of bounds
	if startIdx >= totalItems {
		// Return empty slice of the same type
		emptySlice := reflect.MakeSlice(sliceVal.Type(), 0, 0).Interface()
		return emptySlice, totalItems, nil
	}
	if endIdx > totalItems {
		endIdx = totalItems
	}

	// Create a new slice with the paginated items
	resultSlice := reflect.MakeSlice(sliceVal.Type(), endIdx-startIdx, endIdx-startIdx)

	// Copy the elements to the result slice
	for i := 0; i < endIdx-startIdx; i++ {
		resultSlice.Index(i).Set(sliceVal.Index(startIdx + i))
	}

	return resultSlice.Interface(), totalItems, nil
}
