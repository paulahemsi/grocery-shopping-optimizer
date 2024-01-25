package main

import (
	"encoding/json"
	"fmt"
	"os"
	"math"
)


type userList struct {
	Product []userProducts `json:"products"`
}

type userProducts struct {
	Name	 productName `json:"name"`
	Quantity int    `json:"quantity"`
}

type productData struct {
	Price    int    `json:"price"`
	Quantity int    `json:"quantity"`
}

type marketList struct {
	Fee      int       `json:"fee"`
	Products []product `json:"products"`
}

type product struct {
	Name    productName `json:"name"`
	Data    productData `json:"data"`
}

type itemData struct {
	MarketID int
	Price int
}

type market struct {
	ID int
	Fee int
	Products map[productName]productData
}

type BestPrice struct {
	MarketID marketID
	Price int
}

type marketID int
type productName string

const (
	PRICES_DIFF = "price_diff"
)

func main() {
	getProducts()
}

func getProducts() {
	userList := parseUserListInfos("userGroceriesList.json")
	checkoutItems := make(map[productName]itemData)
	availableMarkets := buildMarketList()
	marketChargedFees := make([]bool, len(availableMarkets))
	marketBestPrices := buildMarketBestPrices()

	for _, item := range userList.Product {
		findBestDeal(item.Name, availableMarkets, marketChargedFees, marketBestPrices, checkoutItems)
	}

	totalPrice := calculateTotalPrice(checkoutItems, marketChargedFees, availableMarkets)
	
	printResults(checkoutItems, totalPrice, availableMarkets, marketChargedFees)
}

func calculateTotalPrice(checkoutItems map[productName]itemData, marketChargedFees []bool, availableMarkets []market) int {
	totalPrice := 0
	for _, item := range checkoutItems {
		totalPrice += item.Price
	}
	totalPrice += addFees(availableMarkets, marketChargedFees)
	return totalPrice
}

//Optimization algorithm

func findBestDeal(itemName productName, markets []market, fees []bool, marketBestPrices []map[productName]int, checkoutItems map[productName]itemData) itemData {
	var winningProductPrice int
	bestDeal := math.MaxInt
	itemInfos := itemData {}
	bestPrice := BestPrice{
		MarketID: -1,
		Price: math.MaxInt,
	}

	for marketId := range markets {
		fee := markets[marketId].Fee
		productPrice := markets[marketId].Products[itemName].Price
		deal := productPrice + fee

		updateBestPrice(productPrice, &bestPrice, marketId)
		updateBestDeal(deal, &bestDeal, &itemInfos, markets, &winningProductPrice, marketId, itemName)
	}

	checkoutItems[itemName] = itemInfos
	processBestPrice(bestPrice, winningProductPrice, itemName, marketBestPrices, fees, checkoutItems, markets)

	if isMarketFirstBuy(fees, itemInfos.MarketID) {
		fees[itemInfos.MarketID] = true
		checkPreviousProducts(checkoutItems, marketBestPrices, itemInfos.MarketID)
	}

	return itemInfos
}

func processBestPrice(bestPrice BestPrice, winningProductPrice int, itemName productName, marketBestPrices []map[productName]int, fees []bool, checkoutItems map[productName]itemData, markets []market) {
	if bestPrice.Price > winningProductPrice {
		return
	}
	marketBestPrices[bestPrice.MarketID][itemName] = bestPrice.Price
	marketBestPrices[bestPrice.MarketID][PRICES_DIFF] += winningProductPrice - bestPrice.Price

	if marketBestPrices[bestPrice.MarketID][PRICES_DIFF] > markets[bestPrice.MarketID].Fee {
		checkPreviousProducts(checkoutItems, marketBestPrices, int(bestPrice.MarketID))
		updateFees(fees, checkoutItems)
	}
}

func checkPreviousProducts(checkoutItems map[productName]itemData, marketBestPrices []map[productName]int, marketId int) {
	for name, price := range marketBestPrices[marketId] {
		if name == PRICES_DIFF || checkoutItems[name].MarketID == marketId {
			continue
		}
		checkoutItems[name] = itemData{
			MarketID: marketId,
			Price:    price,
		}
	}
}

func updateFees(fees []bool, checkoutItems map[productName]itemData) {
	for i := range fees {
		fees[i] = false
	}
	for _, item := range checkoutItems {
		fees[item.MarketID] = true
	}
}

func updateBestPrice(productPrice int, bestPrice *BestPrice, marketId int) {
	if productPrice < bestPrice.Price {
		bestPrice.Price = productPrice
		bestPrice.MarketID = marketID(marketId)
	}
}

func updateBestDeal(deal int, bestDeal *int, itemInfos *itemData, markets []market, winningProductPrice *int, marketId int, itemName productName) {
	if deal < *bestDeal {
		*bestDeal = deal
		itemInfos.MarketID = markets[marketId].ID
		*winningProductPrice = markets[itemInfos.MarketID].Products[itemName].Price
		itemInfos.Price = *winningProductPrice
	}
}

func isMarketFirstBuy(fees []bool, marketID int) bool {
	return !fees[marketID]
}

func addFees(markets []market, fees []bool) int {
	sum := 0
	for i := range fees {
		if fees[i] {
			sum += markets[i].Fee
		}
	}
	return sum
}

//Parser

func buildMarketBestPrices() []map[productName]int {
	marketBestPrices := make([]map[productName]int, 3)
	for i := range marketBestPrices {
		marketBestPrices[i] = make(map[productName]int)
	}
	return marketBestPrices
}

func buildMarketList() []market {
	market1 := parseMarketInfos("Market0.json", 0)
	market2 := parseMarketInfos("Market1.json", 1)
	market3 := parseMarketInfos("Market2.json", 2)
	return []market{market1, market2, market3}
}


func parseMarketInfos(fileName string, marketID int) market {
	var marketData marketList
	marketFile, err := os.Open(fileName)
	defer marketFile.Close()
	check(err)

	jsonParser := json.NewDecoder(marketFile)
	err = jsonParser.Decode(&marketData)
	check(err)

	products := makeMap(marketData.Products)
	return market{
		ID: marketID,
		Fee: marketData.Fee,
		Products: products,
	}
}

func makeMap(list []product) map[productName]productData {
	productsMap := make(map[productName]productData)
	for _, item := range list{
		productsMap[item.Name] = item.Data
	}
	return productsMap
}

func parseUserListInfos(fileName string) userList {
	var user userList
	userFile, err := os.Open(fileName)
	defer userFile.Close()
	check(err)

	jsonParser := json.NewDecoder(userFile)
	err = jsonParser.Decode(&user)
	check(err)

	return user
}

//Auxiliary functions

func printResults(checkoutItems map[productName]itemData, totalPrice int, markets []market, fees []bool) {
	fmt.Println(checkoutItems)
	fmt.Println(fees)
	fmt.Println(totalPrice, "=", checkPrice(checkoutItems, markets, fees))
}

func checkPrice(checkoutItems map[productName]itemData, markets []market, fees []bool) int {
	sum := 0
	for _, infos := range checkoutItems {
		sum += infos.Price
	}
	sum += addFees(markets, fees)
	return sum
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
