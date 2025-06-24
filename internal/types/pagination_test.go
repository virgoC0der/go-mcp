package types

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultPaginationOptions(t *testing.T) {
	defaults := DefaultPaginationOptions()
	assert.Equal(t, 1, defaults.Page)
	assert.Equal(t, 50, defaults.PageSize)
}

func TestCalculateTotalPages(t *testing.T) {
	tests := []struct {
		name       string
		totalItems int
		pageSize   int
		expected   int
	}{
		{"zero items", 0, 10, 0},
		{"exact division", 100, 10, 10},
		{"with remainder", 105, 10, 11},
		{"page size 1", 10, 1, 10},
		{"large page size", 10, 100, 1},
		{"zero page size", 10, 0, 10}, // Should default to page size 1
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := calculateTotalPages(tt.totalItems, tt.pageSize)
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func TestNewPaginatedResult(t *testing.T) {
	items := []string{"a", "b", "c"}
	totalItems := 105
	pageSize := 10
	currentPage := 2
	expectedTotalPages := 11 // Calculated by calculateTotalPages(105, 10)

	result := NewPaginatedResult(items, totalItems, pageSize, currentPage)

	assert.Equal(t, items, result.Items)
	assert.Equal(t, totalItems, result.TotalItems)
	assert.Equal(t, pageSize, result.PageSize)
	assert.Equal(t, currentPage, result.CurrentPage)
	assert.Equal(t, expectedTotalPages, result.TotalPages)
}

func TestApplyPaginationToSlice(t *testing.T) {
	type testStruct struct {
		ID   int
		Name string
	}

	// Prepare test data
	ints := make([]int, 15)
	for i := 0; i < 15; i++ {
		ints[i] = i + 1
	}
	strings := []string{"a", "b", "c", "d", "e"}
	structs := []testStruct{
		{1, "one"},
		{2, "two"},
		{3, "three"},
		{4, "four"},
	}
	emptyInts := []int{}
	nilSlice := []int(nil)

	tests := []struct {
		name          string
		sliceInput    any
		options       PaginationOptions
		expectedSlice any
		expectedCount int
		expectError   bool
	}{
		// Basic cases for ints
		{"int slice page 1", ints, PaginationOptions{Page: 1, PageSize: 5}, []int{1, 2, 3, 4, 5}, 15, false},
		{"int slice page 2", ints, PaginationOptions{Page: 2, PageSize: 5}, []int{6, 7, 8, 9, 10}, 15, false},
		{"int slice page 3 (last full)", ints, PaginationOptions{Page: 3, PageSize: 5}, []int{11, 12, 13, 14, 15}, 15, false},
		{"int slice page 4 (out of bounds)", ints, PaginationOptions{Page: 4, PageSize: 5}, []int{}, 15, false},
		{"int slice last partial page", ints, PaginationOptions{Page: 2, PageSize: 10}, []int{11, 12, 13, 14, 15}, 15, false},

		// Edge cases
		{"page size larger than total", ints, PaginationOptions{Page: 1, PageSize: 20}, ints, 15, false},
		{"page 0", ints, PaginationOptions{Page: 0, PageSize: 5}, []int{1, 2, 3, 4, 5}, 15, false},   // Should default to page 1
		{"page size 0", ints, PaginationOptions{Page: 1, PageSize: 0}, ints[:15], 15, false},         // Should default to default page size (50, but here 15 fits)
		{"page size negative", ints, PaginationOptions{Page: 1, PageSize: -2}, ints[:15], 15, false}, // Should default to default page size

		// Different types
		{"string slice", strings, PaginationOptions{Page: 1, PageSize: 3}, []string{"a", "b", "c"}, 5, false},
		{"struct slice", structs, PaginationOptions{Page: 2, PageSize: 2}, []testStruct{{3, "three"}, {4, "four"}}, 4, false},

		// Empty/nil inputs
		{"empty slice", emptyInts, PaginationOptions{Page: 1, PageSize: 5}, []int{}, 0, false},
		{"nil slice", nilSlice, PaginationOptions{Page: 1, PageSize: 5}, []int(nil), 0, false},

		// Non-slice input
		{"not a slice", 123, PaginationOptions{Page: 1, PageSize: 5}, nil, 0, true},
		{"map input", map[string]int{"a": 1}, PaginationOptions{Page: 1, PageSize: 5}, nil, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			paginatedSlice, totalCount, err := ApplyPaginationToSlice(tt.sliceInput, tt.options)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, paginatedSlice)
				assert.Equal(t, 0, totalCount)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedCount, totalCount)

				// Deep equality check for slices
				assert.True(t, reflect.DeepEqual(tt.expectedSlice, paginatedSlice),
					fmt.Sprintf("Expected slice %v, but got %v", tt.expectedSlice, paginatedSlice))

				// Check type consistency for non-nil results
				if tt.sliceInput != nil && reflect.ValueOf(tt.sliceInput).Kind() == reflect.Slice && reflect.ValueOf(tt.sliceInput).Len() > 0 && !tt.expectError {
					assert.Equal(t, reflect.TypeOf(tt.sliceInput), reflect.TypeOf(paginatedSlice))
				}
			}
		})
	}
}
