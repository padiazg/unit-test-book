package benchmark_tests

import (
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConcatJoin(t *testing.T) {
	assert.Equal(t, "a,b,c", ConcatJoin([]string{"a", "b", "c"}))
}

func TestConcatPlus(t *testing.T) {
	assert.Equal(t, "a,b,c", ConcatPlus([]string{"a", "b", "c"}))
}

func TestConcatBuilder(t *testing.T) {
	assert.Equal(t, "a,b,c", ConcatBuilder([]string{"a", "b", "c"}))
}

func TestSortIntsCopy(t *testing.T) {
	nums := []int{3, 1, 2}
	result := SortIntsCopy(nums)
	assert.Equal(t, []int{1, 2, 3}, result)
	assert.Equal(t, []int{3, 1, 2}, nums) // original unchanged
}

func TestSortIntsInPlace(t *testing.T) {
	nums := []int{3, 1, 2}
	result := SortIntsInPlace(nums)
	assert.Equal(t, []int{1, 2, 3}, result)
	assert.Equal(t, []int{1, 2, 3}, nums) // original modified
}

func TestSortUsersByAge(t *testing.T) {
	users := []User{{Name: "Alice", Age: 30}, {Name: "Bob", Age: 20}}
	sorted := SortUsersByAge(users)
	require.Len(t, sorted, 2)
	assert.Equal(t, "Bob", sorted[0].Name)
	assert.Equal(t, "Alice", sorted[1].Name)
}

func TestParseCSVLine(t *testing.T) {
	assert.Equal(t, []string{"a", "b"}, ParseCSVLine("a,b"))
	assert.Nil(t, ParseCSVLine(""))
}

func TestFormatUser(t *testing.T) {
	assert.Equal(t, "Alice (30)", FormatUser(User{Name: "Alice", Age: 30}))
}

func TestFilterAdults(t *testing.T) {
	users := []User{
		{Name: "Alice", Age: 30},
		{Name: "Bob", Age: 15},
		{Name: "Charlie", Age: 18},
	}
	adults := FilterAdults(users)
	assert.Len(t, adults, 2)
}

func TestSumInts(t *testing.T) {
	assert.Equal(t, 15, SumInts([]int{1, 2, 3, 4, 5}))
	assert.Equal(t, 0, SumInts(nil))
}

func BenchmarkConcatJoin(b *testing.B) {
	items := []string{"alpha", "beta", "gamma", "delta", "epsilon"}
	for i := 0; i < b.N; i++ {
		ConcatJoin(items)
	}
}

func BenchmarkConcatPlus(b *testing.B) {
	items := []string{"alpha", "beta", "gamma", "delta", "epsilon"}
	for i := 0; i < b.N; i++ {
		ConcatPlus(items)
	}
}

func BenchmarkConcatBuilder(b *testing.B) {
	items := []string{"alpha", "beta", "gamma", "delta", "epsilon"}
	for i := 0; i < b.N; i++ {
		ConcatBuilder(items)
	}
}

func BenchmarkSortIntsCopy(b *testing.B) {
	nums := []int{9, 8, 7, 6, 5, 4, 3, 2, 1, 0}
	for i := 0; i < b.N; i++ {
		SortIntsCopy(nums)
	}
}

func BenchmarkSortIntsInPlace(b *testing.B) {
	nums := []int{9, 8, 7, 6, 5, 4, 3, 2, 1, 0}
	for i := 0; i < b.N; i++ {
		SortIntsInPlace(nums)
	}
}

func BenchmarkParseCSVLine(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ParseCSVLine("alpha,beta,gamma,delta,epsilon")
	}
}

func BenchmarkFilterAdults(b *testing.B) {
	users := make([]User, 1000)
	for i := range users {
		users[i] = User{Name: "User", Age: i % 30}
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		FilterAdults(users)
	}
}

func BenchmarkSumInts(b *testing.B) {
	nums := make([]int, 10000)
	for i := range nums {
		nums[i] = i
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		SumInts(nums)
	}
}

func BenchmarkSumIntsRange(b *testing.B) {
	nums := make([]int, 10000)
	for i := range nums {
		nums[i] = i
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		SumIntsRange(nums)
	}
}

func BenchmarkSortUsersByAge(b *testing.B) {
	users := make([]User, 1000)
	for i := range users {
		users[i] = User{Name: "User", Age: 1000 - i}
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		SortUsersByAge(users)
	}
}

func BenchmarkLargeConcat(b *testing.B) {
	items := make([]string, 1000)
	for i := range items {
		items[i] = "value"
	}

	b.Run("Join", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			strings.Join(items, ",")
		}
	})

	b.Run("Builder", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			var bld strings.Builder
			for j, s := range items {
				if j > 0 {
					bld.WriteString(",")
				}
				bld.WriteString(s)
			}
			sink = bld.String()
		}
	})
}

func BenchmarkSortStdLib(b *testing.B) {
	nums := []int{3, 1, 4, 1, 5, 9, 2, 6, 5, 3}

	b.Run("sort.Ints", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			tmp := make([]int, len(nums))
			copy(tmp, nums)
			sort.Ints(tmp)
		}
	})

	b.Run("sort.Slice", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			tmp := make([]int, len(nums))
			copy(tmp, nums)
			sort.Slice(tmp, func(i, j int) bool { return tmp[i] < tmp[j] })
		}
	})
}
