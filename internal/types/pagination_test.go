package types

import (
	"testing"
)

func TestDefaultPaginationOptions(t *testing.T) {
	options := DefaultPaginationOptions()

	if options.Page != 1 {
		t.Errorf("Expected default page to be 1, got %d", options.Page)
	}

	if options.PageSize != 50 {
		t.Errorf("Expected default page size to be 50, got %d", options.PageSize)
	}
}

func TestNewPaginatedResult(t *testing.T) {
	// Create test data
	items := []string{"item1", "item2", "item3"}
	totalItems := 10
	pageSize := 3
	currentPage := 2

	// Create paginated result
	result := NewPaginatedResult(items, totalItems, pageSize, currentPage)

	// Check fields
	if result.TotalItems != totalItems {
		t.Errorf("TotalItems mismatch: expected %d, got %d", totalItems, result.TotalItems)
	}

	if result.PageSize != pageSize {
		t.Errorf("PageSize mismatch: expected %d, got %d", pageSize, result.PageSize)
	}

	if result.CurrentPage != currentPage {
		t.Errorf("CurrentPage mismatch: expected %d, got %d", currentPage, result.CurrentPage)
	}

	// Expected total pages = ceil(10/3) = 4
	expectedTotalPages := 4
	if result.TotalPages != expectedTotalPages {
		t.Errorf("TotalPages mismatch: expected %d, got %d", expectedTotalPages, result.TotalPages)
	}

	// Check Items
	resultItems, ok := result.Items.([]string)
	if !ok {
		t.Fatalf("Items is not of expected type []string")
	}

	if len(resultItems) != len(items) {
		t.Errorf("Items length mismatch: expected %d, got %d", len(items), len(resultItems))
	}
}

func TestCalculateTotalPages(t *testing.T) {
	testCases := []struct {
		totalItems    int
		pageSize      int
		expectedPages int
	}{
		{10, 3, 4},  // 10/3 = 3.33, ceil to 4
		{10, 5, 2},  // 10/5 = 2
		{10, 10, 1}, // 10/10 = 1
		{0, 5, 0},   // 0/5 = 0
		{5, 0, 5},   // Division by zero handled by setting pageSize to 1
		{11, 3, 4},  // 11/3 = 3.67, ceil to 4
	}

	for _, tc := range testCases {
		result := calculateTotalPages(tc.totalItems, tc.pageSize)
		if result != tc.expectedPages {
			t.Errorf("calculateTotalPages(%d, %d): expected %d, got %d",
				tc.totalItems, tc.pageSize, tc.expectedPages, result)
		}
	}
}

func TestApplyPaginationToSlice(t *testing.T) {
	// Create test data - a slice of integers
	testData := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	testCases := []struct {
		name          string
		data          interface{}
		options       PaginationOptions
		expectedItems []int
		expectedTotal int
		expectedError bool
	}{
		{
			name:          "First page",
			data:          testData,
			options:       PaginationOptions{Page: 1, PageSize: 3},
			expectedItems: []int{1, 2, 3},
			expectedTotal: 10,
			expectedError: false,
		},
		{
			name:          "Middle page",
			data:          testData,
			options:       PaginationOptions{Page: 2, PageSize: 3},
			expectedItems: []int{4, 5, 6},
			expectedTotal: 10,
			expectedError: false,
		},
		{
			name:          "Last complete page",
			data:          testData,
			options:       PaginationOptions{Page: 3, PageSize: 3},
			expectedItems: []int{7, 8, 9},
			expectedTotal: 10,
			expectedError: false,
		},
		{
			name:          "Partial last page",
			data:          testData,
			options:       PaginationOptions{Page: 4, PageSize: 3},
			expectedItems: []int{10},
			expectedTotal: 10,
			expectedError: false,
		},
		{
			name:          "Empty result (page beyond end)",
			data:          testData,
			options:       PaginationOptions{Page: 5, PageSize: 3},
			expectedItems: []int{},
			expectedTotal: 10,
			expectedError: false,
		},
		{
			name:          "Invalid page number",
			data:          testData,
			options:       PaginationOptions{Page: 0, PageSize: 3},
			expectedItems: []int{1, 2, 3},
			expectedTotal: 10,
			expectedError: false,
		},
		{
			name:          "Invalid page size",
			data:          testData,
			options:       PaginationOptions{Page: 1, PageSize: 0},
			expectedItems: []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}[:DefaultPaginationOptions().PageSize], // Should use default page size
			expectedTotal: 10,
			expectedError: false,
		},
		{
			name:          "Empty slice",
			data:          []int{},
			options:       PaginationOptions{Page: 1, PageSize: 3},
			expectedItems: []int{},
			expectedTotal: 0,
			expectedError: false,
		},
		{
			name:          "Non-slice input",
			data:          123, // Not a slice
			options:       PaginationOptions{Page: 1, PageSize: 3},
			expectedItems: nil,
			expectedTotal: 0,
			expectedError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, total, err := ApplyPaginationToSlice(tc.data, tc.options)

			// Check error
			if tc.expectedError && err == nil {
				t.Errorf("Expected an error but got none")
			}
			if !tc.expectedError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}

			// If we expect an error, don't check the results
			if tc.expectedError {
				return
			}

			// Check total
			if total != tc.expectedTotal {
				t.Errorf("Total items mismatch: expected %d, got %d", tc.expectedTotal, total)
			}

			// Check result items
			if result != nil {
				resultItems, ok := result.([]int)
				if !ok && len(tc.expectedItems) > 0 {
					t.Fatalf("Result is not of expected type []int")
				}

				if len(resultItems) != len(tc.expectedItems) {
					t.Errorf("Items length mismatch: expected %d, got %d", len(tc.expectedItems), len(resultItems))
				}

				for i, v := range tc.expectedItems {
					if i < len(resultItems) && resultItems[i] != v {
						t.Errorf("Item at index %d mismatch: expected %d, got %d", i, v, resultItems[i])
					}
				}
			}
		})
	}
}
