package main

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"sort"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ingredient struct {
	Name   string  `bson:"Name"`   // Name of the ingredient
	PDH    float64 `bson:"PDH"`    // Potion Direct Healing
	DHM    float64 `bson:"DHM"`    // DH Multiplier
	PDP    float64 `bson:"PDP"`    // Potion Direct Poison
	DPM    float64 `bson:"DPM"`    // DP Multiplier
	PHoT   float64 `bson:"PHoT"`   // Potion Healing Over Time
	HoTM   float64 `bson:"HoTM"`   // HoT Multiplier
	PHL    float64 `bson:"PHL"`    // Potion Healing Length
	HLM    float64 `bson:"HLM"`    // HoT Length Multiplier
	PPoT   float64 `bson:"PPoT"`   // Potion Poison Over Time
	PoTM   float64 `bson:"PoTM"`   // PoT Multiplier
	PPL    float64 `bson:"PPL"`    // Potion Poison Length
	PLM    float64 `bson:"PLM"`    // Poison Length Multiplier
	PAlc   float64 `bson:"PAlc"`   // Potion Alcohol
	AlcM   float64 `bson:"AlcM"`   // Alcohol Multiplier
	LM     float64 `bson:"LM"`     // Lore Multiplier  (1 + F * L / 100) where F is lore factor and L is the lore level
	Weight bool    `bson:"Weight"` // Weight of the ingredient
}

// ingredientStack
type ingredientStack struct {
	Ingredient ingredient `bson:"Ingredient"` // The ingredient
	Amount     float64    `bson:"Amount"`     // Amount of the ingredient in the potion (maximum 10000)
}

// potion
type potion struct {
	IngredientStacks []ingredientStack `bson:"IngredientStacks"`
}

type bestDHPotion struct {
	Potion potion  `bson:"Potion"`
	DH     float64 `bson:"DH"`
	Type   string  `bson:"Type"`
}

const A float64 = 1.2 // Advanced Potion Making (APM) multiplier ( for 100)
// Ci - Count of the ingredient in the potion
// N - Total number of ingredients (sum of all counts)
// Bi - The base property value of the ingredient
// Mi - The multiplier value of the ingredient
// Li - Lore multiplier of the ingredient

// const MaxIngredientAmount = 10000
// const MaxPotionWeight = 8000
// const MaxIngredientStacks = 16

// Вычисление DH для ингридиента
func (ingredient ingredient) CalculateDH(N float64) float64 {
	res := A * (ingredient.LM * ingredient.PDH * (N / N)) * (1 + ingredient.DHM*math.Sqrt(N/N))
	return res
}

// Вычисление DH для зелья
func (potion potion) CalculateDH() float64 {
	potionAmount := 0.0

	for _, ingredientStack := range potion.IngredientStacks {
		potionAmount += ingredientStack.Amount
	}

	DH := 0.0
	for _, ingredientStack := range potion.IngredientStacks {
		DH += ingredientStack.Ingredient.LM * ingredientStack.Ingredient.PDH * (ingredientStack.Amount / potionAmount)
	}

	DM := 1.0
	for _, ingredientStack := range potion.IngredientStacks {
		DM *= 1 + ingredientStack.Ingredient.DHM*math.Sqrt(ingredientStack.Amount/potionAmount)
	}

	res := A * DH * DM
	return res
}

// Вычисление веса зелья
func (potion potion) CalculateWeight() float64 { // максимальный вес 8000
	potionAmount := 0.0
	for _, ingredientStack := range potion.IngredientStacks {
		if ingredientStack.Ingredient.Weight {
			potionAmount += ingredientStack.Amount
		}
	}
	return potionAmount
}

// Получение списка ингридиентов из базы данных
func TakeIngridients(Client *mongo.Database) []ingredient {
	IngridColl := Client.Collection("Reagents")
	corsor, err := IngridColl.Find(context.TODO(), bson.D{})
	if err != nil {
		panic(err)
	}

	ingridients := []ingredient{}

	if err = corsor.All(context.TODO(), &ingridients); err != nil {
		panic(err)
	}

	return ingridients
}

func TakeBestPotion(Client *mongo.Database, Type string) bestDHPotion {
	PotionColl := Client.Collection("Potions")
	var result bestDHPotion
	filter := bson.D{{"Type", Type}}
	err := PotionColl.FindOne(context.TODO(), filter).Decode(&result)

	if err != nil {
		panic(err)
	}
	return result
}

func WhriteBestPotion(Client *mongo.Database, result bestDHPotion) {
	PotionColl := Client.Collection("Potions")
	_, err := PotionColl.InsertOne(context.TODO(), result)
	if err != nil {
		panic(err)
	}
}

func UpdateBestPotion(Client *mongo.Database, result bestDHPotion) (*mongo.UpdateResult, error) {
	PotionColl := Client.Collection("Potions")

	filter := bson.D{{"Type", result.Type}}
	replacement := bestDHPotion{Potion: result.Potion, DH: result.DH, Type: result.Type}

	return PotionColl.ReplaceOne(context.TODO(), filter, replacement)
}

func InitDatabase() *mongo.Client {
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI("mongodb+srv://mrtimholl:vDuvVusbbHdUQUqC@cluster0.uylufal.mongodb.net/?retryWrites=true&w=majority&appName=Cluster0").SetServerAPIOptions(serverAPI)

	client, err := mongo.Connect(context.TODO(), opts)
	if err != nil {
		panic(err)
	}

	return client
}

func GenerateRandomPotion(ingredients []ingredient, MaxIngredientStacks int, MaxIngredientAmount int, MaxPotionWeight float64) potion {
	//fmt.Println("Generating random potion...")
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	p := potion{IngredientStacks: []ingredientStack{}}
	used := make(map[string]bool) // New map to track used ingredients.
	for _, u := range ingredients {
		used[u.Name] = false
	}

	for len(p.IngredientStacks) < MaxIngredientStacks {
		ingredientIndex := r.Intn(len(ingredients))
		// Check if the ingredient is already used.
		if used[ingredients[ingredientIndex].Name] {
			var count_of_unused = 0
			for _, u := range ingredients {
				if !used[u.Name] {
					count_of_unused++
				}
			}
			if count_of_unused == 0 {
				break
			}
			continue // Skip this ingredient if it's already used.
		}
		amount := float64(r.Intn(MaxIngredientAmount))
		newStack := ingredientStack{
			Ingredient: ingredients[ingredientIndex],
			Amount:     amount,
		}

		p.IngredientStacks = append(p.IngredientStacks, newStack)

		if p.CalculateWeight() > MaxPotionWeight {
			p.IngredientStacks = p.IngredientStacks[:len(p.IngredientStacks)-1]
			continue
		}
		used[ingredients[ingredientIndex].Name] = true // Mark this ingredient as used.
	}

	return p
}

func Crossover(parent1, parent2 potion, ingredients []ingredient) potion {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	//fmt.Println("Crossover")
	// Initialize an empty child potion.
	child := parent1
	used := make(map[string]bool) // New map to track used ingredients.

	minParent := int(math.Min(float64(len(parent1.IngredientStacks)), float64(len(parent2.IngredientStacks))))
	var crossoverPoint1, crossoverPoint2 = 0, 0
	if minParent > 1 {
		// Выбираем две случайные точки кроссовера в диапазоне ингредиентов
		crossoverPoint1 = r.Intn(minParent)
		crossoverPoint2 = r.Intn(minParent)

		// Убедимся, что crossoverPoint1 меньше crossoverPoint2
		if crossoverPoint1 > crossoverPoint2 {
			crossoverPoint1, crossoverPoint2 = crossoverPoint2, crossoverPoint1
		}

		// fmt.Println("crossoverPoint1: ", crossoverPoint1, "crossoverPoint2: ", crossoverPoint2)
		// fmt.Println("parent1: ", parent1)
		// fmt.Println("parent2: ", parent2)

		// Копируем ингредиенты от первого родителя до первой точки кроссовера
		for i := 0; i < crossoverPoint1; i++ {
			if !used[parent1.IngredientStacks[i].Ingredient.Name] {
				child.IngredientStacks[i] = parent1.IngredientStacks[i]
				used[parent1.IngredientStacks[i].Ingredient.Name] = true
			}

		}

		//fmt.Println("child after parent1: ", child)

		// Копируем ингредиенты от второго родителя между двумя точками кроссовера
		for i := crossoverPoint1; i < crossoverPoint2; i++ {
			if !used[parent2.IngredientStacks[i].Ingredient.Name] {
				child.IngredientStacks[i] = parent2.IngredientStacks[i]
				used[parent2.IngredientStacks[i].Ingredient.Name] = true
			}
		}

		//fmt.Println("child after parent2: ", child)

		// Копируем оставшиеся ингредиенты от первого родителя после второй точки кроссовера
		for i := crossoverPoint2; i < minParent; i++ {
			// fmt.Println("i: ", i)
			// fmt.Println("parent1.IngredientStacks[i].Ingredient.Name: ", parent1.IngredientStacks[i].Ingredient.Name)
			if !used[parent1.IngredientStacks[i].Ingredient.Name] {
				child.IngredientStacks[i] = parent1.IngredientStacks[i]
				used[parent1.IngredientStacks[i].Ingredient.Name] = true
			}
		}

		//fmt.Println("child after parent1: ", child)
	}

	return child
}

// Mutate вносит случайные изменения в зелье
func Mutate(potion *potion, ingredients []ingredient, MaxIngredientAmount int, MaxPotionWeight float64) {
	//fmt.Println("Mutate")
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	mutationRateAmount := 0.1      // 10% вероятность мутации количества
	mutationRateIngredient := 0.05 // 5% вероятность мутации ингредиента

	// Проходим по всем ингредиентам зелья
	for i := range potion.IngredientStacks {
		mutationAmount := 0.0
		InAnount := potion.IngredientStacks[i].Amount
		inIgridient := potion.IngredientStacks[i].Ingredient
		// Мутация количества ингредиента
		if r.Float64() < mutationRateAmount {
			// Выбираем новое количество ингредиента
			mutationAmount = float64(r.Intn(MaxIngredientAmount))
			potion.IngredientStacks[i].Amount = mutationAmount
		}

		// Мутация самого ингредиента
		if r.Float64() < mutationRateIngredient {
			// Выбираем случайный ингредиент из списка доступных
			randomIndex := r.Intn(len(ingredients))
			potion.IngredientStacks[i].Ingredient = ingredients[randomIndex]
		}

		// Убедимся, что вес зелья не превышает максимально допустимый после мутации
		// Если вес зелья превышает максимально допустимый, то отменяем мутацию
		if potion.CalculateWeight() > MaxPotionWeight {
			potion.IngredientStacks[i].Amount = InAnount
			potion.IngredientStacks[i].Ingredient = inIgridient
		}
	}
}

// Генетичесский алгоритм поиска оптимального зелья
// Инициализируется начальная популяция случайно сгенерированных зелий.
// Осуществляется оценка каждого зелья по функции CalculateDH и сортировка популяции по убыванию DH.
// Выбираются лучшие конфигурации для скрещивания.
// Лучшие конфигурации используются для создания нового поколения зелий посредством скрещивания и мутаций.
// Процесс повторяется на протяжении заданного числа поколений или пока не будет найдена оптимальная конфигурация.
// Возвращается зелье с лучшим значением DH.
func OptimizePotionDH(ingredients []ingredient, MaxIngredientAmount int, MaxPotionWeight float64, MaxIngredientStacks int) potion {
	//fmt.Println("OptimizePotionDH")
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	populationSize := 1000
	generations := 5000
	selectionSize := 100

	// Инициализация начальной популяции
	population := make([]potion, populationSize)
	for i := range population {
		population[i] = GenerateRandomPotion(ingredients, MaxIngredientStacks, MaxIngredientAmount, MaxPotionWeight)
	}

	var bestPotion potion
	bestDH := -1.0

	for generation := 0; generation < generations; generation++ {
		// Оценка
		sort.Slice(population, func(i, j int) bool {
			return population[i].CalculateDH() > population[j].CalculateDH()
		})

		// fmt.Println("Generation: ", generation)
		// fmt.Println("Population: ", population)
		// fmt.Println("Best potion: ", population[0].CalculateDH())

		// Проверяем лучшее текущее решение
		if bestDH < population[0].CalculateDH() {
			bestDH = population[0].CalculateDH()
			bestPotion = population[0]
		}

		// Селекция
		selected := population[:selectionSize]

		// Создание нового поколения
		for i := selectionSize; i < populationSize; i++ {
			// Скрещивание (можно реализовать функцию Crossover)
			parent1 := selected[r.Intn(selectionSize)]
			parent2 := selected[r.Intn(selectionSize)]
			child := Crossover(parent1, parent2, ingredients)

			// Мутация (можно реализовать функцию Mutate)
			Mutate(&child, ingredients, MaxIngredientAmount, MaxPotionWeight)

			population[i] = child
		}
	}

	return bestPotion
}

func StartSearch(DataBase *mongo.Database, Type string, MaxIngredientAmount int, MaxPotionWeight float64, MaxIngredientStacks int) {
	fmt.Println("StartSearch: ", Type)
	for {
		fmt.Println("StartIteration: ", Type)
		ingridients := TakeIngridients(DataBase)

		BestPotion := TakeBestPotion(DataBase, Type)

		potion := OptimizePotionDH(ingridients, MaxIngredientAmount, MaxPotionWeight, MaxIngredientStacks)
		fmt.Println("New potion "+Type+": ", potion, " DH: ", potion.CalculateDH())
		if BestPotion.DH < potion.CalculateDH() {
			NewBest := bestDHPotion{
				DH:     potion.CalculateDH(),
				Potion: potion,
				Type:   Type,
			}
			UpdateBestPotion(DataBase, NewBest)
			fmt.Println(Type, " new best: ", NewBest)
		}
	}
}

func main() {

	Client := InitDatabase()
	defer func() {
		if err := Client.Disconnect(context.TODO()); err != nil {
			panic(err)
		}
	}()
	DataBase := Client.Database("Alchemy")
	fmt.Println("Connected to MongoDB!")

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		StartSearch(DataBase, "DH8000", 10000, 8000, 16)
	}()

	go func() {
		defer wg.Done()
		StartSearch(DataBase, "DH40", 100, 40, 16)
	}()

	wg.Wait()

	fmt.Println("Done?")

}
