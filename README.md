# Go Weighted Random

A Go library for efficient weighted random selection using Vose's Alias Method.

## Performance

| Dimension | Complexity |
| --- | ---: |
| Time | O(1) |
| Memory | O(n) |

## Installation
```bash
$ go get github.com/nikole-dunixi/go-weighted-random
```

## Example Usage
You will need to initialize an instance of `rand.Random`, or any other implementation that satisfies the local interface `RandIntN`, and any items that need to be selected.

There is no imposed limitations around concurrency, [except for those needed to support the standard library `rand.Random`](https://stackoverflow.com/a/48959983/1478636).

```go
rand := rand.New(rand.NewSource(time.Now().Unix()))
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
```