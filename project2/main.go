package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strconv"
	"time"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const Time = 500 * time.Millisecond
const	Size       = 10


var companies = []string{"AMZ", "FLP", "SP", "HYN", "AZO"}


func main() {
	r := gin.Default()
	r.GET("/categories/:category/products", getProducts)
	r.GET("/categories/:category/products/:productid", getDetails)
	fmt.Println("Server started at port : 9876")
	if err := r.Run(":9876"); err != nil {
		log.Fatal("Failed to run server: ", err)
	}
}

type Product struct {
	ID          string  `json:"id"`
	Name        string  `json:"productName"`
	Category    string  `json:"category"`
	Company     string  `json:"company"`
	Price       float64 `json:"price"`
	Rating      float64 `json:"rating"`
	Discount    float64 `json:"discount"`
	Availability string `json:"availability"`
	Description string  `json:"description"`
}

type ProductResponse struct {
	Products []Product `json:"products"`
}

func GetAllProducts(company, category string, minPrice, maxPrice float64, limit int) ([]Product, error) {
	client := http.Client{Timeout: Time}
	url := "http://20.244.56.144/test/companies/" + company + "/categories/" + category + "/products/top-" + strconv.Itoa(limit) + "?minPrice=" + strconv.FormatFloat(minPrice, 'f', 2, 64) + "&maxPrice=" + strconv.FormatFloat(maxPrice, 'f', 2, 64)
	res, err := client.Get(url)
	if err != nil {
		log.Println("Error getting products:", err)
		return nil, err
	}
	defer res.Body.Close()

	var items ProductResponse
	if err := json.NewDecoder(res.Body).Decode(&items); err != nil {
		log.Println("Error getting responses:", err)
		return nil, err
	}

	for i,_ := range items.Products {
		items.Products[i].ID = uuid.New().String()
		items.Products[i].Company = company
		if items.Products[i].Discount > 0 {
			items.Products[i].Availability = "yes"
		} else {
			items.Products[i].Availability = "out-of-stock"
		}
	}

	return items.Products, nil
}

func FinalProducts(productsList [][]Product, sortBy string, order string) []Product {
	var allProducts []Product
	for _, products := range productsList {
		allProducts = append(allProducts, products...)
	}

	sort.SliceStable(allProducts, func(i, j int) bool {
		switch sortBy {
		case "price":
			if order == "asc" {
				return allProducts[i].Price < allProducts[j].Price
			}
			return allProducts[i].Price > allProducts[j].Price
		case "rating":
			if order == "asc" {
				return allProducts[i].Rating < allProducts[j].Rating
			}
			return allProducts[i].Rating > allProducts[j].Rating
		case "discount":
			if order == "asc" {
				return allProducts[i].Discount < allProducts[j].Discount
			}
			return allProducts[i].Discount > allProducts[j].Discount
		case "company":
			if order == "asc" {
				return allProducts[i].Company < allProducts[j].Company
			}
			return allProducts[i].Company > allProducts[j].Company
		default:
			return true
		}
	})

	return allProducts
}

func getProducts(c *gin.Context) {
	category := c.Param("category")
	minPrice, _ := strconv.ParseFloat(c.DefaultQuery("minPrice", "0"), 64)
	maxPrice, _ := strconv.ParseFloat(c.DefaultQuery("maxPrice", "1000000"), 64)
	n, _ := strconv.Atoi(c.DefaultQuery("n", "10"))
	sortBy := c.DefaultQuery("sortBy", "price")
	order := c.DefaultQuery("order", "asc")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))

	var productsList [][]Product
	for _, company := range companies {
		products, err := GetAllProducts(company, category, minPrice, maxPrice, n)
		if err == nil {
			productsList = append(productsList, products)
		}
	}

	allProducts := FinalProducts(productsList, sortBy, order)

	start := (page - 1) * Size
	end := start + Size
	if start >= len(allProducts) {
		c.JSON(http.StatusOK, gin.H{
			"products":        []Product{},
			"currentPage":     page,
			"totalPages":      (len(allProducts) +Size - 1) / Size,
			"totalProducts":   len(allProducts),
			"productsPerPage": Size,
		})
		return
	}
	if end > len(allProducts) {
		end = len(allProducts)
	}

	c.JSON(http.StatusOK, gin.H{
		"products":        allProducts[start:end],
		"currentPage":     page,
		"totalPages":      (len(allProducts) + Size - 1) / Size,
		"totalProducts":   len(allProducts),
		"productsPerPage": Size,
	})
}

func getDetails(c *gin.Context) {
	category := c.Param("category")
	product_ID := c.Param("productid")

	dummyProduct := Product{
		ID:          product_ID,
		Name:        "Laptop 1",
		Category:    category,
		Company:     "Dummy Company",
		Price:       2236,
		Rating:      4.7,
		Discount:    63,
		Availability: "yes",
		Description: "This is a dummy product.",
	}

	c.JSON(http.StatusOK, dummyProduct)
}
