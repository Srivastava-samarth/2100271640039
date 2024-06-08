package main

import (
	"encoding/json"
	"net/http"
	"time"
	"log"
	"sort"
	"strconv"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	testServerURL  = "http://20.244.56.144/test/companies/"
	requestTimeout = 500 * time.Millisecond
	pageSize       = 10
)

var (
	companies = []string{"AMZ", "FLP", "SP", "HYN", "AZO"}
)

func main() {
	r := gin.Default()
	r.GET("/categories/:category/products", getProductsHandler)
	r.GET("/categories/:category/products/:productid", getProductDetailsHandler)
	if err := r.Run(":8080"); err != nil {
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

type TestServerResponse struct {
	Products []Product `json:"products"`
}

func fetchProducts(company, category string, minPrice, maxPrice float64, limit int) ([]Product, error) {
	client := http.Client{Timeout: requestTimeout}
	url := testServerURL + company + "/categories/" + category + "/products/top-" + strconv.Itoa(limit) + "?minPrice=" + strconv.FormatFloat(minPrice, 'f', 2, 64) + "&maxPrice=" + strconv.FormatFloat(maxPrice, 'f', 2, 64)
	resp, err := client.Get(url)
	if err != nil {
		log.Println("Error fetching products:", err)
		return nil, err
	}
	defer resp.Body.Close()

	var result TestServerResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Println("Error decoding response:", err)
		return nil, err
	}

	for i := range result.Products {
		result.Products[i].ID = uuid.New().String()
		result.Products[i].Company = company
		if result.Products[i].Discount > 0 {
			result.Products[i].Availability = "yes"
		} else {
			result.Products[i].Availability = "out-of-stock"
		}
	}

	return result.Products, nil
}

func mergeAndSortProducts(productsList [][]Product, sortBy string, order string) []Product {
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

func getProductsHandler(c *gin.Context) {
	category := c.Param("category")
	minPrice, _ := strconv.ParseFloat(c.DefaultQuery("minPrice", "0"), 64)
	maxPrice, _ := strconv.ParseFloat(c.DefaultQuery("maxPrice", "1000000"), 64)
	n, _ := strconv.Atoi(c.DefaultQuery("n", "10"))
	sortBy := c.DefaultQuery("sortBy", "price")
	order := c.DefaultQuery("order", "asc")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))

	var productsList [][]Product
	for _, company := range companies {
		products, err := fetchProducts(company, category, minPrice, maxPrice, n)
		if err == nil {
			productsList = append(productsList, products)
		}
	}

	mergedProducts := mergeAndSortProducts(productsList, sortBy, order)

	start := (page - 1) * pageSize
	end := start + pageSize
	if start >= len(mergedProducts) {
		c.JSON(http.StatusOK, gin.H{
			"products":        []Product{},
			"currentPage":     page,
			"totalPages":      (len(mergedProducts) + pageSize - 1) / pageSize,
			"totalProducts":   len(mergedProducts),
			"productsPerPage": pageSize,
		})
		return
	}
	if end > len(mergedProducts) {
		end = len(mergedProducts)
	}

	c.JSON(http.StatusOK, gin.H{
		"products":        mergedProducts[start:end],
		"currentPage":     page,
		"totalPages":      (len(mergedProducts) + pageSize - 1) / pageSize,
		"totalProducts":   len(mergedProducts),
		"productsPerPage": pageSize,
	})
}

func getProductDetailsHandler(c *gin.Context) {
	category := c.Param("category")
	productID := c.Param("productid")
	dummyProduct := Product{
		ID:          productID,
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
