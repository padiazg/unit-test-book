package benchmark_tests

import (
	"fmt"
	"sort"
	"strings"
)

type User struct {
	Name string
	Age  int
}

func ConcatJoin(items []string) string {
	return strings.Join(items, ",")
}

func ConcatPlus(items []string) string {
	var out strings.Builder
	for _, s := range items {
		out.WriteString(s + ",")
	}
	if len(out.String()) > 0 {
		return out.String()[:len(out.String())-1]
	}
	return out.String()
}

func ConcatBuilder(items []string) string {
	var b strings.Builder
	for i, s := range items {
		if i > 0 {
			b.WriteString(",")
		}
		b.WriteString(s)
	}
	return b.String()
}

func SortIntsCopy(nums []int) []int {
	out := make([]int, len(nums))
	copy(out, nums)
	sort.Ints(out)
	return out
}

func SortIntsInPlace(nums []int) []int {
	sort.Ints(nums)
	return nums
}

func SortUsersByAge(users []User) []User {
	out := make([]User, len(users))
	copy(out, users)
	sort.Slice(out, func(i, j int) bool {
		return out[i].Age < out[j].Age
	})
	return out
}

func ParseCSVLine(line string) []string {
	if line == "" {
		return nil
	}
	return strings.Split(line, ",")
}

func FormatUser(u User) string {
	return fmt.Sprintf("%s (%d)", u.Name, u.Age)
}

func FilterAdults(users []User) []User {
	out := make([]User, 0, len(users))
	for _, u := range users {
		if u.Age >= 18 {
			out = append(out, u)
		}
	}
	return out
}

func SumInts(nums []int) int {
	total := 0
	for _, n := range nums {
		total += n
	}
	return total
}

func SumIntsRange(nums []int) int {
	total := 0
	for i := range nums {
		total += nums[i]
	}
	return total
}
