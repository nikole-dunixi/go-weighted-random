package weightedrand

import (
	"fmt"
	"slices"
	"strings"

	"github.com/shopspring/decimal"
)

var one decimal.Decimal

func init() {
	one = decimal.NewFromInt(1)
}

// Weight is a type constraint that allows any signed or unsigned integer type.
// It is intended for use in generic functions or types that operate on weighted values,
// where the weight can be represented by any integer type.
type Weight interface {
	// integers
	int | int8 | int16 | int32 | int64 |
		// unsigned integers
		uint | uint8 | uint16 | uint32 | uint64 |
		// support for decimal.Decimal itself
		decimal.Decimal
}

// WeightedRandom is a generic interface that defines a method for selecting
// the next value of type T based on weighted randomness. Implementations of
// this interface should provide logic to return a value of type T according
// to their specific weighted random selection algorithm.
type WeightedRandom[T any] interface {
	Next() T
}

// RandIntN defines an interface for random number generators that can produce
// non-negative pseudo-random integers less than a specified value. It provides
// methods for generating both int and int64 values within a given range.
type RandIntN interface {
	Intn(n int) int
	Int63n(n int64) int64
}

// WeightedItem represents an item with an associated weight.
// TItem is the type of the item, and TWeight is the type of the weight (which must satisfy the Weight constraint).
// This struct is typically used in weighted random selection algorithms.
type WeightedItem[TItem any, TWeight Weight] struct {
	Item   TItem
	Weight TWeight
}

type voseAliasMethodRandom[TItem any] struct {
	random RandIntN
	tuples []aliasTuple[TItem]
}

type weightedItem[TItem any] struct {
	Item   TItem
	Weight decimal.Decimal
}

func (item WeightedItem[TItem, TWeight]) String() string {
	return fmt.Sprintf(
		"{weight: %d, item: %v}",
		item.Weight,
		item.Item,
	)
}

type aliasTuple[TItem any] struct {
	probability decimal.Decimal
	primaryItem TItem
	aliasedItem *TItem
}

func (tuple aliasTuple[TItem]) String() string {
	aliasString := "[nil]"
	if tuple.aliasedItem != nil {
		aliasString = fmt.Sprintf("%v", *tuple.aliasedItem)
	}
	return fmt.Sprintf(
		"{probability: %s, primary: %v, alias: %s}",
		tuple.probability.String(),
		tuple.primaryItem,
		aliasString,
	)
}

// NewAliasVoseMethod constructs a new WeightedRandom instance using the Alias Method (Vose's algorithm)
// for efficient weighted random sampling. It takes a random number generator and a variadic list of
// WeightedItem values, and returns a WeightedRandom implementation that allows O(1) sampling.
//
// The function panics if no items are provided.
//
// Type Parameters:
//   - TItem:   The type of the items to be sampled.
//   - TWeight: The type representing the weight of each item.
//
// Parameters:
//   - random: A RandIntN implementation used for random number generation.
//   - items:  A variadic list of WeightedItem values, each containing an item and its associated weight.
//
// Returns:
//   - WeightedRandom[TItem]: An implementation that supports efficient weighted random selection.
//
// Panics:
//   - If no items are provided or weights are negative.
//
// Example usage:
//
//	wr := NewAliasVoseMethod(randSource, WeightedItem{Item: "A", Weight: 2}, WeightedItem{Item: "B", Weight: 3})
func NewAliasVoseMethod[TItem any, TWeight Weight](random RandIntN, items ...WeightedItem[TItem, TWeight]) WeightedRandom[TItem] {
	if len(items) == 0 {
		panic("at least one item must be provided")
	}
	// Create two worklists, Small and Large.
	small, large := createPartitionedItems(items)

	// Create slices alias and prob, each of size n
	tuples := make([]aliasTuple[TItem], 0, len(items))
	for ; len(small) > 0 && len(large) > 0; small, large = small[1:], large[1:] {
		lesser, greater := small[0], large[0]
		// Using the smaller probability, create the alias for the two items.
		tuples = append(tuples,
			aliasTuple[TItem]{
				probability: lesser.Weight,
				primaryItem: lesser.Item,
				aliasedItem: &greater.Item,
			},
		)
		// Take the larger probability and find how much is "remaining" when
		// you take the two into consideration.
		nextItem := weightedItem[TItem]{
			Item:   greater.Item,
			Weight: greater.Weight.Add(lesser.Weight).Sub(one),
		}
		if nextProbability := nextItem.Weight; nextProbability.LessThan(one) {
			small = append(small, nextItem)
		} else {
			large = append(large, nextItem)
		}
	}
	// For all remaining large items, place them into their own singlular
	// aliases.
	for ; len(large) > 0; large = large[1:] {
		greaterItem := large[0]
		tuples = append(tuples,
			aliasTuple[TItem]{
				probability: one,
				primaryItem: greaterItem.Item,
			},
		)
	}
	// For all remaining small items, place them into their own singlular
	// aliases.
	for ; len(small) > 0; small = small[1:] {
		smallerItem := small[0]
		tuples = append(tuples,
			aliasTuple[TItem]{
				probability: one,
				primaryItem: smallerItem.Item,
			},
		)
	}
	return voseAliasMethodRandom[TItem]{
		random: random,
		tuples: tuples,
	}
}

func createPartitionedItems[TValue any, TWeight Weight](items []WeightedItem[TValue, TWeight]) ([]weightedItem[TValue], []weightedItem[TValue]) {
	// Create intermediate list to ensure we don't modify the user's
	// input.
	itemBuffer := make([]weightedItem[TValue], 0, len(items))
	totalWeight := decimal.Zero
	// First pass through the slice creates the duplicate slice
	// and sums the total weight
	for _, currentItem := range items {
		// If no weight is provided, it is assumed to be 1
		currentWeight := weightAsDecimal(currentItem.Weight)
		if currentWeight.Equal(decimal.Zero) {
			currentWeight = one
		} else if currentWeight.LessThan(decimal.Zero) {
			panic(fmt.Sprintf("weight must be non-negative value, but was %s", currentWeight.String()))
		}

		totalWeight = totalWeight.Add(currentWeight)
		itemBuffer = append(itemBuffer, weightedItem[TValue]{
			Item:   currentItem.Item,
			Weight: currentWeight,
		})
	}
	// Second pass through the slice normalizes the probabilities
	// and makes them relative to each other.
	itemCount := decimal.NewFromUint64(uint64(len(itemBuffer)))
	for i := range itemBuffer {
		currentItem := itemBuffer[i]
		replacementWeight := currentItem.Weight.
			Mul(itemCount).
			Div(totalWeight)
		currentItem.Weight = replacementWeight
		itemBuffer[i] = currentItem
	}
	// Sort the items. Find the index of the first item that is >= 1.
	// Use the index to create sub-slices.
	slices.SortFunc(itemBuffer, func(a, b weightedItem[TValue]) int {
		return a.Weight.Cmp(b.Weight)
	})
	index := slices.IndexFunc(itemBuffer, func(item weightedItem[TValue]) bool {
		return item.Weight.GreaterThanOrEqual(one)
	})

	// Copy into dedicated slices. We cannot optimize with subslices, because
	// we may append items into the list as they are processed.
	bufferSmall := itemBuffer[:index]
	bufferLarge := itemBuffer[index:]
	resultSmall := make([]weightedItem[TValue], len(bufferSmall))
	resultLarge := make([]weightedItem[TValue], len(bufferLarge))
	copy(resultSmall, bufferSmall)
	copy(resultLarge, bufferLarge)
	return resultSmall, resultLarge
}

func weightAsDecimal[TWeight Weight](value TWeight) decimal.Decimal {
	switch value := any(value).(type) {
	case int:
		// int will have at least 32 bits, but is not an alias
		// for int32; per lanuage standards
		return decimal.NewFromInt(int64(value))
	case int8:
		return decimal.NewFromInt32(int32(value))
	case int16:
		return decimal.NewFromInt32(int32(value))
	case int32:
		return decimal.NewFromInt32(int32(value))
	case int64:
		return decimal.NewFromInt(value)
	case uint:
		// uint will have at least 32 bits, but is not an alias
		// for uint32; per lanuage standards
		return decimal.NewFromUint64(uint64(value))
	case uint8:
		return decimal.NewFromUint64(uint64(value))
	case uint16:
		return decimal.NewFromUint64(uint64(value))
	case uint32:
		return decimal.NewFromUint64(uint64(value))
	case uint64:
		return decimal.NewFromUint64(value)
	case decimal.Decimal:
		// If we have a decimal already, we just return it back
		return value
	default:
		panic(fmt.Sprintf("unsupported numerical value %d (%T)", value, value))
	}
}

func (aliasMethod voseAliasMethodRandom[TItem]) Next() TItem {
	// First, perform a fair dice roll.
	fairDiceRoll := aliasMethod.random.Intn(len(aliasMethod.tuples))
	fairlyChosenTuple := aliasMethod.tuples[fairDiceRoll]
	// Second, perform an unfair dice roll.
	max := int64(100)
	unfairCoinToss := decimal.NewFromInt(aliasMethod.random.Int63n(max)).
		Div(decimal.NewFromInt(max))
	if unfairCoinToss.LessThan(fairlyChosenTuple.probability) {
		return fairlyChosenTuple.primaryItem
	}
	return *fairlyChosenTuple.aliasedItem
}

func (aliasMethod voseAliasMethodRandom[TItem]) String() string {
	randomString := fmt.Sprintf("%T", aliasMethod.random)
	tupleStrings := make([]string, 0, len(aliasMethod.tuples))
	for item := range slices.Values(aliasMethod.tuples) {
		tupleStrings = append(tupleStrings, item.String())
	}
	return fmt.Sprintf(
		"{random: %s, tuples: [%s]}",
		randomString, strings.Join(tupleStrings, ", "),
	)
}
