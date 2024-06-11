package main

import (
	"fmt"
	"math"
)

type ingredient struct {
	name string
	PDH  float64 // Potion Direct Healing
	DHM  float64 // DH Multiplier
	PDP  float64 // Potion Direct Poison
	DPM  float64 // DP Multiplier
	PHoT float64 // Potion Healing Over Time
	HoTM float64 // HoT Multiplier
	PHL  float64 // Potion Healing Length
	HLM  float64 // HoT Length Multiplier
	PPoT float64 // Potion Poison Over Time
	PoTM float64 // PoT Multiplier
	PPL  float64 // Potion Poison Length
	PLM  float64 // Poison Length Multiplier
	PAlc float64 // Potion Alcohol
	AlcM float64 // Alcohol Multiplier
	LM   float64 // Lore Multiplier  (1 + F * L / 100) where F is lore factor and L is the lore level
}

type ingredientStack struct {
	ingredient ingredient
	Amount     float64
}

type potion struct {
	ingredientStacks []ingredientStack
}

const A float64 = 1.2 // Advanced Potion Making (APM) multiplier ( for 100)
// Ci - Count of the ingredient in the potion
// N - Total number of ingredients (sum of all counts)
// Bi - The base property value of the ingredient
// Mi - The multiplier value of the ingredient
// Li - Lore multiplier of the ingredient

func (ingredient ingredient) CalculateDH(N float64) float64 {
	res := A * (ingredient.LM * ingredient.PDH * (N / N)) * (1 + ingredient.DHM*math.Sqrt(N/N))
	return res
}

func (potion potion) CalculateDH() float64 {
	potionAmount := 0.0

	for _, ingredientStack := range potion.ingredientStacks {
		potionAmount += ingredientStack.Amount
	}

	DH := 0.0
	for _, ingredientStack := range potion.ingredientStacks {
		DH += ingredientStack.ingredient.LM * ingredientStack.ingredient.PDH * (ingredientStack.Amount / potionAmount)
	}

	DM := 1.0
	for _, ingredientStack := range potion.ingredientStacks {
		DM *= 1 + ingredientStack.ingredient.DHM*math.Sqrt(ingredientStack.Amount/potionAmount)
	}

	res := A * DH * DM
	return res

}

func main() {

	Sea_Dew_Leaves := ingredient{
		PDH: 1.2,
		LM:  (5.0 / 3.0),
	}

	Muse_Fruit := ingredient{
		PDH: 0.15,
		DHM: 0.44,
		LM:  (5.0 / 3.0),
	}

	fmt.Println(Sea_Dew_Leaves)
	fmt.Println(Sea_Dew_Leaves.CalculateDH(11))
	fmt.Println(Muse_Fruit.CalculateDH(11))

	potion := potion{
		ingredientStacks: []ingredientStack{
			{ingredient: Sea_Dew_Leaves, Amount: 11},
			{ingredient: Muse_Fruit, Amount: 1},
		},
	}
	fmt.Println(potion.CalculateDH())

}
