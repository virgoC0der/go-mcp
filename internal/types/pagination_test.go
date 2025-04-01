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
