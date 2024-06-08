package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	Time =500 * time.Millisecond
	size = 10
)

var numbersWindow = []int{}

var validIds = map[string]string{"p": "primes", "f": "fibo", "e": "even", "r": "rand"};

func main() {
	port := os.Getenv("PORT")
	if port == ""{
		port = "9876"
	}
	router := gin.New();
	router.GET("/numbers/:id",getFinalNumbers);
	fmt.Println("Starting the server on port 9876....")
	router.Run(":" + port);
}

type ResponseFromTestServer struct{
	Numbers []int `json:"numbers"`
}

func getNumbers(id string) ([]int,error){
	client := http.Client{Timeout: Time}
	res,err := client.Get("http://20.244.56.144/test/"+ id);
	if err!=nil{
		log.Println("Error in fetching the numbers",err);
	}
	defer res.Body.Close();
	var result ResponseFromTestServer;
	if err:= json.NewDecoder(res.Body).Decode(&result);
	err!=nil{
		log.Println("Error in getting response",err);
	}
	return result.Numbers,nil;
}

func getWindowSize(response []int){
	numberSet := make(map[int]bool)
	for _,num := range numbersWindow{
		numberSet[num] = true
	}
	for _,num := range response{
		if !numberSet[num]{
			if len(numbersWindow) >= 10{
				numbersWindow = numbersWindow[1:]
			}
			numbersWindow = append(numbersWindow,num);
			numberSet[num] = true;
		}
	}
}


func getAverage() float64{
	if len(numbersWindow) == 0{
		return 0
	}
	sum := 0
	for _,num := range(numbersWindow){
		sum+=num
	}
	return float64(sum) / float64(len(numbersWindow))
}

func getFinalNumbers(c *gin.Context){
	id := c.Param("id");
	ID,valid := validIds[id]
	if !valid{
		c.JSON(http.StatusBadRequest,gin.H{"error":"Invalid Id"})
		return
	}
	response,err := getNumbers(ID)
	if err!=nil{
		c.JSON(http.StatusInternalServerError,gin.H{"error":"Failed to fetch the numbers"})
		return
	}

	previousState := make([]int,len(numbersWindow))
	copy(previousState,numbersWindow);
	getWindowSize(response)
	currentState := make([]int,len(numbersWindow))
	copy(currentState,numbersWindow);
	average:=getAverage();


	c.JSON(http.StatusOK,gin.H{
		"fetchedNumbers": response,
		"previousState" :previousState,
		"currentState" : currentState,
		"average" : average,
	})
}
