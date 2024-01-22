package main

import (
	"encoding/json"
	"fmt"
	"os"
	"math"
)

const (
	MARKET_1_ID = 1
	MARKET_2_ID = 2
	MARKET_3_ID = 3
)

func main() {
	getProducts()
}

type userList struct {
	Product []userProducts `json:"products"`
}

type userProducts struct {
	Name	 string `json:"name"`
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
	Name    string `json:"name"`
	Data    productData `json:"data"`
}

type checkoutList struct {
	Price int
	Items []itemData
}

type itemData struct {
	Name string
	MarketID int
	Quantity int
}

type market struct {
	ID int
	Fee int
	Products map[string]productData
}

//TODO: considerar quantidades
func getProducts() {
	market1 := parseMarketInfos("Market1.json", MARKET_1_ID)
	market2 := parseMarketInfos("Market2.json", MARKET_2_ID)
	market3 := parseMarketInfos("Market3.json", MARKET_3_ID)
	availableMarkets := []market{market1, market2, market3}

	userList := parseUserListInfos("userGroceriesList.json")

	var checkout checkoutList
	marketFees := make([]bool, len(availableMarkets) + 1)

	for _, item := range userList.Product {
		itemInfos, price := findBestDeal(item, availableMarkets, marketFees)
		checkout.Items = append(checkout.Items, itemInfos)
		checkout.Price += price
	}
	fmt.Println(checkout)
	fmt.Println(checkout.Price, checkPrice(checkout, availableMarkets, marketFees))
}

func checkPrice(checkout checkoutList, markets []market, fees []bool) int {
	sum := 0
	for _, item := range checkout.Items {
		sum += markets[item.MarketID - 1].Products[item.Name].Price
	}
	for i := range fees {
		if fees[i] {
			sum += markets[i - 1].Fee
		}
	}
	return sum
}

func findBestDeal(item userProducts, markets []market, fees []bool) (itemData, int) {
	bestDeal := math.MaxInt
	itemInfos := itemData {
		Name: item.Name,
	}
	for i := 0; i < len(markets); i++ {
		fee := 0
		if !fees[markets[i].ID] {
			fee = markets[i].Fee
		}
		deal := markets[i].Products[item.Name].Price + fee
		if deal < bestDeal {
			bestDeal = deal
			itemInfos.MarketID = markets[i].ID
		}
	}
	fees[itemInfos.MarketID] = true
	return itemInfos, bestDeal
}

func makeMap(list []product) map[string]productData {
	productsMap := make(map[string]productData)
	for _, item := range list{
		productsMap[item.Name] = item.Data
	}
	return productsMap
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

func check(err error) {
	if err != nil {
		panic(err)
	}
}
