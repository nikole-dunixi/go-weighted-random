package weightedrand

import (
	"fmt"
	"math/rand"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const tolerance = 0.05

type MarbleColor string

const (
	Red    MarbleColor = `RED`
	Orange MarbleColor = `ORANGE`
	Yellow MarbleColor = `YELLOW`
	Green  MarbleColor = `GREEN`
	Blue   MarbleColor = `BLUE`
)

type MarbleColorCounts map[MarbleColor]int64

func (mcc MarbleColorCounts) String() string {
	items := make([]string, 0, len(mcc))
	for color, count := range mcc {
		items = append(items, fmt.Sprintf("%s=%d", color, count))
	}
	sort.Strings(items)
	return "{counts: " + strings.Join(items, ", ") + "}"
}

func BenchmarkWeightedRand(b *testing.B) {

	permutations := map[string][]WeightedItem[MarbleColor, uint]{
		"1:1": {
			WeightedItem[MarbleColor, uint]{
				Item:   Red,
				Weight: 1,
			},
			WeightedItem[MarbleColor, uint]{
				Item:   Orange,
				Weight: 1,
			},
		},
		"1:1:10": {
			WeightedItem[MarbleColor, uint]{
				Item:   Red,
				Weight: 1,
			},
			WeightedItem[MarbleColor, uint]{
				Item:   Orange,
				Weight: 1,
			},
			WeightedItem[MarbleColor, uint]{
				Item:   Yellow,
				Weight: 10,
			},
		},
		"1:10:100": {
			WeightedItem[MarbleColor, uint]{
				Item:   Red,
				Weight: 1,
			},
			WeightedItem[MarbleColor, uint]{
				Item:   Orange,
				Weight: 10,
			},
			WeightedItem[MarbleColor, uint]{
				Item:   Yellow,
				Weight: 100,
			},
		},
		"1:50:100": {
			WeightedItem[MarbleColor, uint]{
				Item:   Red,
				Weight: 1,
			},
			WeightedItem[MarbleColor, uint]{
				Item:   Orange,
				Weight: 50,
			},
			WeightedItem[MarbleColor, uint]{
				Item:   Yellow,
				Weight: 100,
			},
		},
		"1:50:100:1000": {
			WeightedItem[MarbleColor, uint]{
				Item:   Red,
				Weight: 1,
			},
			WeightedItem[MarbleColor, uint]{
				Item:   Orange,
				Weight: 50,
			},
			WeightedItem[MarbleColor, uint]{
				Item:   Yellow,
				Weight: 100,
			},
			WeightedItem[MarbleColor, uint]{
				Item:   Green,
				Weight: 1000,
			},
		},
	}

	for _, iterations := range []uint{100, 1000, 100_000, 10_000_000} {
		benchmarkWeightedRand(b, uint(iterations), permutations)
	}
}

func benchmarkWeightedRand(
	b *testing.B, iterations uint, permutations map[string][]WeightedItem[MarbleColor, uint],
) {
	b.Run(fmt.Sprintf("iterations %d", iterations), func(b *testing.B) {
		for name, items := range permutations {
			b.Run(name, func(b *testing.B) {
				r := rand.New(rand.NewSource(time.Now().Unix()))
				wr := NewAliasVoseMethod(r, items...)
				for range iterations {
					_ = wr.Next()
				}
			})
		}
	})
}

func TestWeightedRand(t *testing.T) {
	t.Run("panic", func(t *testing.T) {
		t.Run("no items", func(t *testing.T) {
			assert.Panics(t, func() {
				NewAliasVoseMethod[int, int](nil)
			})
		})
		t.Run("items with negative weight", func(t *testing.T) {
			testPanicsWithNegativeWeight[int](t, -1)
			testPanicsWithNegativeWeight[int8](t, -1)
			testPanicsWithNegativeWeight[int16](t, -1)
			testPanicsWithNegativeWeight[int32](t, -1)
			testPanicsWithNegativeWeight[int64](t, -1)
			testPanicsWithNegativeWeight[decimal.Decimal](t, decimal.NewFromInt(-1))
		})
	})
	t.Run("items with weights", func(t *testing.T) {
		testWeightedProbabilitiesWithinTolerance(t,
			"1(unweighted):3", []WeightedItem[MarbleColor, uint]{
				{
					Item: Blue,
				},
				{
					Item:   Red,
					Weight: 3,
				},
			})
		testWeightedProbabilitiesWithinTolerance(t,
			"1:3", []WeightedItem[MarbleColor, uint]{
				{
					Item:   Blue,
					Weight: 1,
				},
				{
					Item:   Red,
					Weight: 3,
				},
			})

		testWeightedProbabilitiesWithinTolerance(t,
			"1(unweighted):1(unweighted)", []WeightedItem[MarbleColor, uint]{
				{
					Item: Blue,
				},
				{
					Item: Red,
				},
			})
		testWeightedProbabilitiesWithinTolerance(t,
			"1(unweighted):1", []WeightedItem[MarbleColor, uint]{
				{
					Item: Blue,
				},
				{
					Item:   Red,
					Weight: 1,
				},
			})
		testWeightedProbabilitiesWithinTolerance(t,
			"1:1(unweighted)", []WeightedItem[MarbleColor, uint]{
				{
					Item:   Blue,
					Weight: 1,
				},
				{
					Item: Red,
				},
			})
		testWeightedProbabilitiesWithinTolerance(t,
			"1:1", []WeightedItem[MarbleColor, uint]{
				{
					Item:   Blue,
					Weight: 1,
				},
				{
					Item:   Red,
					Weight: 1,
				},
			})
		testWeightedProbabilitiesWithinTolerance(t,
			"1:1:1", []WeightedItem[MarbleColor, uint]{
				{
					Item:   Blue,
					Weight: 1,
				},
				{
					Item:   Red,
					Weight: 1,
				},
				{
					Item:   Yellow,
					Weight: 1,
				},
			})
		testWeightedProbabilitiesWithinTolerance(t,
			"1:1:3", []WeightedItem[MarbleColor, uint]{
				{
					Item:   Blue,
					Weight: 1,
				},
				{
					Item:   Red,
					Weight: 1,
				},
				{
					Item:   Yellow,
					Weight: 3,
				},
			})

		testWeightedProbabilitiesWithinTolerance(t,
			"1(unweighted):1(unweighted):3", []WeightedItem[MarbleColor, uint]{
				{
					Item: Blue,
				},
				{
					Item: Red,
				},
				{
					Item:   Yellow,
					Weight: 3,
				},
			})
		testWeightedProbabilitiesWithinTolerance(t,
			"1:5:100", []WeightedItem[MarbleColor, uint]{
				{
					Item:   Blue,
					Weight: 1,
				},
				{
					Item:   Red,
					Weight: 5,
				},
				{
					Item:   Yellow,
					Weight: 100,
				},
			})
		testWeightedProbabilitiesWithinTolerance(t,
			"1:5:15:100", []WeightedItem[MarbleColor, uint]{
				{
					Item:   Blue,
					Weight: 1,
				},
				{
					Item:   Red,
					Weight: 5,
				},
				{
					Item:   Yellow,
					Weight: 15,
				},
				{
					Item:   Green,
					Weight: 100,
				},
			})
	})
}

func testWeightedProbabilitiesWithinTolerance(
	t *testing.T, name string, items []WeightedItem[MarbleColor, uint],
) {
	t.Helper()
	const iterations int64 = 100_000

	totalWeight := decimal.Zero
	for _, item := range items {
		if currentWeight := weightAsDecimal(item.Weight); decimal.Zero.Equal(currentWeight) {
			totalWeight = totalWeight.Add(one)
		} else {
			totalWeight = totalWeight.Add(currentWeight)
		}
	}

	expectedProportions := make(map[MarbleColor]decimal.Decimal)
	for _, item := range items {
		if currentWeight := weightAsDecimal(item.Weight); decimal.Zero.Equal(currentWeight) {
			expectedProportions[item.Item] = one.Div(totalWeight)
		} else {
			expectedProportions[item.Item] = currentWeight.Div(totalWeight)
		}
	}

	t.Run(name, func(t *testing.T) {
		r := rand.New(rand.NewSource(time.Now().Unix()))
		wr := NewAliasVoseMethod(r, items...)

		counts := make(MarbleColorCounts)
		for range iterations {
			marbleColor := wr.Next()
			counts[marbleColor] += 1
		}

		for color, count := range counts {
			actualProportion := decimal.NewFromInt(count).Div(decimal.NewFromInt(iterations))
			expectedProportion := expectedProportions[color]
			actualDifferenceDecimal := actualProportion.Sub(expectedProportion).Abs()
			toleranceDecimal := decimal.NewFromFloat(tolerance)

			assert.Truef(t,
				actualDifferenceDecimal.LessThanOrEqual(toleranceDecimal),
				"the proportion %s was not within %s tolerance of %s (was %s)",
				actualProportion.String(), toleranceDecimal.String(), expectedProportion.String(), actualDifferenceDecimal.String(),
			)
		}
	})
}

func testPanicsWithNegativeWeight[TWeight Weight](t *testing.T, weight TWeight) {
	t.Helper()
	testname := fmt.Sprintf("%T", weight)
	t.Run(testname, func(t *testing.T) {
		assert.Panics(t, func() {
			NewAliasVoseMethod(nil, WeightedItem[string, TWeight]{
				Item:   testname,
				Weight: weight,
			})
		})
	})
}

func FixtureDecimal(t *testing.T, v string) decimal.Decimal {
	t.Helper()
	result, err := decimal.NewFromString(v)
	require.NoErrorf(t, err, "testcase had invalid value for expected decimal: %s", v)
	return result
}
