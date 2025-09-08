package weightedrand_test

import (
	"fmt"
	"math/rand"

	"github.com/nikole-dunixi/weightedrand"
)

func ExampleNewAliasVoseMethod() {
	rand := rand.New(rand.NewSource(1337))

	wr := weightedrand.NewAliasVoseMethod(rand,
		weightedrand.WeightedItem[string, int]{
			Item:   "Hollow Knight: Silksong",
			Weight: 1,
		},
		weightedrand.WeightedItem[string, int]{
			Item:   "Don't Starve Together",
			Weight: 3,
		},
		weightedrand.WeightedItem[string, int]{
			Item:   "Stardew Valley",
			Weight: 3,
		},
		weightedrand.WeightedItem[string, int]{
			Item:   "Deep Rock Galactic",
			Weight: 7,
		},
	)

	for range 5 {
		fmt.Printf("%s\n", wr.Next())
	}
	// Output:
	//
	// Stardew Valley
	// Deep Rock Galactic
	// Deep Rock Galactic
	// Stardew Valley
	// Deep Rock Galactic
}
