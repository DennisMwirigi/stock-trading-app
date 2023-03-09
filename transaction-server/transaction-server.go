package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// mock db, actual requests will be sent to a Mongo DB
type account struct {
	ID      string  `json:"id"`
	Balance float64 `json:"balance"`
}

var accounts = []account{
	{ID: "1", Balance: 100},
	{ID: "2", Balance: 200},
	{ID: "3", Balance: 300},
}

type holding struct {
	symbol   string
	quantity float64
	pps      float64
}

type balanceDif struct {
	ID     string
	Amount float64
}

type users struct {
	user_id string
}

// Not used.  There is supposed to be a way to read mongo db stuff directly into struct, but I coldnt get it to work.
type c_bal struct {
	cash_balance int32
}

type quote struct {
	Stock string
	Price float64
	CKey  string // Crytohraphic key
	// add timeout property
}

type order struct {
	ID     string
	Stock  string
	Buy    float64 // amount
	Buy_id int
	// figure out timeout feature
}

var orders = []order{}

func connectDb(databaseUri string) (*mongo.Client, error) {
	// adapted from https://github.com/mongodb/mongo-go-driver/blob/d957e67225a9ea82f1c7159020b4f9fd7c8d441a/README.md#usage
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return mongo.Connect(ctx, options.Client().ApplyURI(databaseUri))
}

// main
func main() {
	router := gin.Default() // initializing Gin router
	router.SetTrustedProxies(nil)

	var db *mongo.Database
	router.Use(func(ctx *gin.Context) {
		ctx.Set("db", db)
		ctx.Next()
	})


	router.GET("/users", getAll) // Do we even need?? Not really

	router.GET("/users/:id", getAccount)

	//router.POST("/newuser", addAccount) Migh be used if we do sign up

	router.PUT("/users/:id/add/:addBal", addBalance)

	router.GET("/users/:id/quote/:stock", getQuote)

	router.POST("/users/:id/buy/:stock/amount/:quantity", buyStock)

	router.POST("/users/:id/sell/:stock/amount/:quantity", sellStock)

	router.GET("/health", healthcheck)

	bind := flag.String("bind", "localhost:8080", "host:port to listen on")
	flag.Parse()

	databaseUri, found := os.LookupEnv("DATABASE_URI")
	if !found {
		log.Fatalln("No DATABASE_URI")
	}

	mongoClient, err := connectDb(databaseUri)
	if err != nil {
		log.Fatalln(err)
	}

	db = mongoClient.Database("daytrading")

	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := mongoClient.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()

	CLI()

	err = router.Run(*bind)
	log.Fatal(err)
}

func getAll(c *gin.Context) {
	// Bad on performance
	r := readMany("users", bson.D{})
	c.IndentedJSON(http.StatusOK, r)
}

func getAccount(c *gin.Context) {
	id := c.Param("id")

	fmt.Println(id)
	r := readOne("users", bson.D{{"user_id", id}})
	n := bson.D{{"none", "none"}}

	if !reflect.DeepEqual(r, n) {
		c.IndentedJSON(http.StatusOK, r)
		return
	}
	// If account not found

	err := insert("users", bson.D{{"user_id", id}})
	if err != "ok" {
		panic(err)
	}
	c.IndentedJSON(http.StatusOK, "success")
}

func addAccount(id string) account {
	// NOT CURRENTLY USED
	var newAccount account
	newAccount.ID = id
	newAccount.Balance = 0

	accounts = append(accounts, newAccount)
	return newAccount
}

// THIS CODE MIGHT BE USEFUL IF WE DO SIGN UP FEATURE
// func addAccount(c *gin.Context) {
// 	var newAccount account

// 	// Call BindJSON to bind the received JSON to newAccount.
// 	if err := c.BindJSON(&newAccount); err != nil {
// 		return
// 	}
// 	// Add the new account to the slice.
// 	accounts = append(accounts, newAccount)
// 	c.IndentedJSON(http.StatusCreated, newAccount)
// }

func addBalance(c *gin.Context) {
	id := c.Param("id")
	bal, err := strconv.Atoi(c.Param("addBal"))
	if err != nil {
		panic(err)
	}

	fmt.Println(id)
	fmt.Println(bal)

	//id := c.Param("id")
	r := updateOne("users", bson.D{{"user_id", id}}, bson.D{{"cash_balance", bal}}, "$inc")

	if r != "ok" {
		panic(r)
	}

}

func getQuote(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, "INCOMPLETE")

}
func getQuoteLocal(sym string) float64 {
	// WILL BE DELETED LATER
	// JUST SO THAT THERE IS A RETURN VALUE
	return 1
}

func getQuoteTEMP(sym string, username string) (string, string, string) {
	//TEMPORARY NAME BECAUSE IT INTERFERS WITH GET QUOTE HTTP METHOD
	//make connection to server
	strEcho := sym + " " + username + "\n"
	servAddr := "quoteserve.seng.uvic.ca:4444"

	tcpAddr, err := net.ResolveTCPAddr("tcp", servAddr)
	if err != nil {
		fmt.Println("\nResolveTCPAddr error: ", err)
		os.Exit(1)
	}

	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		fmt.Println("\nDialTCP error: ", err)
		os.Exit(1)
	}

	//write to server SYM being requested and user
	_, err = conn.Write([]byte(strEcho))
	if err != nil {
		fmt.Println("\nWrite error: ", err)
		os.Exit(1)
	}

	//reading from server
	_reply := make([]byte, 1024)

	_, err = conn.Read(_reply)
	if err != nil {
		fmt.Println("\nRead error: ", err)
		os.Exit(1)
	}

	//parsing reply from server
	reply := strings.Split(strings.ReplaceAll(string(_reply), "\n", ""), ",")
	quotePrice := reply[0]
	timestamp := reply[3]
	cryptKey := reply[4]

	conn.Close()

	return quotePrice, timestamp, cryptKey
}

func buyStock(c *gin.Context) {
	id := c.Param("id")
	stock := c.Param("stock")
	quantity, err := strconv.ParseFloat(c.Param("quantity"), 64)
	pps := getQuoteLocal(stock)

	if err != nil {
		panic("ERR")
	}

	r := readField("users", bson.D{{"user_id", id}}, bson.D{{"cash_balance", 1}})

	n := bson.D{{"none", "none"}}

	if reflect.DeepEqual(r, n) {
		panic("ERROR")
	}
	cost := pps * quantity

	fmt.Printf("\nTYPE = %T\n", r[0][1].Value)
	fmt.Printf("\nTYPE = %T\n", cost)

	// Check if user has enough balance
	switch v := r[0][1].Value.(type) {
	case float64:
		{
			fmt.Println("FLOATING")
			if v > cost {
				r := updateOne("users", bson.D{{"user_id", id}}, bson.D{{"cash_balance", v - cost}}, "$set")
				i := updateOne("users", bson.D{{"user_id", id}}, bson.D{{"account_holdings", bson.D{{"symbol", stock}, {"quantity", quantity}, {"pps", pps}}}}, "$push")
				if i != "ok" {
					panic("PUSH ERROR")
				}
				if r != "ok" {
					panic(r)
				}
				//c.IndentedJSON(http.StatusBadRequest, accounts[index])
				return
			}
		}
	case int64:
		{
			a := float64(v)
			fmt.Println(a)
			fmt.Println(cost)
			if a > cost {
				fmt.Println("YES")
				r := updateOne("users", bson.D{{"user_id", id}}, bson.D{{"cash_balance", a - cost}}, "$set")
				i := updateOne("users", bson.D{{"user_id", id}}, bson.D{{"account_holdings", bson.D{{"symbol", stock}, {"quantity", quantity}, {"pps", pps}}}}, "$push")
				if i != "ok" {
					panic("PUSH ERROR")
				}
				if r != "ok" {
					panic(r)
				}
				//c.IndentedJSON(http.StatusBadRequest, accounts[index])
				return
			}
		}
	case int32:
		{
			a := float64(v)
			fmt.Println(a)
			fmt.Println(cost)
			if a > cost {
				fmt.Println("INT32")
				r := updateOne("users", bson.D{{"user_id", id}}, bson.D{{"cash_balance", a - cost}}, "$set")
				i := updateOne("users", bson.D{{"user_id", id}}, bson.D{{"account_holdings", bson.D{{"symbol", stock}, {"quantity", quantity}, {"pps", pps}}}}, "$push")
				if i != "ok" {
					panic("PUSH ERROR")
				}
				if r != "ok" {
					panic(r)
				}
				//c.IndentedJSON(http.StatusBadRequest, accounts[index])
				return
			}
		}
	}

	// User has enough balance, proceed creating order
	//buy_id := len(orders) + 1
	//newOrder.Buy_id = buy_id
	//orders = append(orders, newOrder)
	//return
	//c.IndentedJSON(http.StatusOK, newOrder)
}

func sellStock(c *gin.Context) {
	id := c.Param("id")
	stock := c.Param("stock")
	quantity, err := strconv.ParseFloat(c.Param("quantity"), 64)
	pps := getQuoteLocal(stock)

	if err != nil {
		panic("ERR")
	}
	y := rawreadField("users", bson.D{{"user_id", id}}, bson.D{{"cash_balance", 1}})

	fmt.Println(y)

	r := rawreadField("users", bson.D{{"user_id", id}}, bson.D{{"account_holdings", 1}})
	n := bson.D{{"none", "none"}}

	if reflect.DeepEqual(r, n) {
		panic("ERROR")
	}
	//fmt.Println("~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~")
	//fmt.Println(r[0][1].Value)

	//fmt.Println()
	//fmt.Printf("TYPE = %T\n\n", r[0][1].Value)
	//fmt.Printf("TYPE = %T\n\n", r[0][0].Value)

	var these_holdings []holding

	//var temp holding

	switch v := r[0][1].Value.(type) {
	case bson.A:
		{
			// Only works with account holdings
			these_holdings = mongo_read_bsonA(v)
		}
	}

	//value of the trade
	value := pps * quantity

	// Check if user has the correct holdings
	for _, holding := range these_holdings {
		if holding.symbol == stock {
			//Will rewrite later
			fmt.Println("TRUE")
			r := updateOne("users", bson.D{{"user_id", id}}, bson.D{{"cash_balance", value}}, "$inc")
			i := updateOne("users", bson.D{{"user_id", id}}, bson.D{{"account_holdings", bson.D{{"symbol", holding.symbol}, {"quantity", holding.quantity}, {"pps", holding.pps}}}}, "$pull")
			if i != "ok" {
				panic("PUSH ERROR")
			}
			f := updateOne("users", bson.D{{"user_id", id}}, bson.D{{"account_holdings", bson.D{{"symbol", stock}, {"quantity", holding.quantity - quantity}, {"pps", pps}}}}, "$push")

			if f != "ok" {
				panic("PUSH ERROR")
			}
			if r != "ok" {
				panic(r)
			}
			//c.IndentedJSON(http.StatusBadRequest, accounts[index])
			return
		}

	}

	// User has enough balance, proceed creating order
	//buy_id := len(orders) + 1
	//newOrder.Buy_id = buy_id
	//orders = append(orders, newOrder)
	//return
	//c.IndentedJSON(http.StatusOK, newOrder)
}

func healthcheck(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	db := c.MustGet("db").(*mongo.Database)
	err := db.Client().Ping(ctx, readpref.SecondaryPreferred())

	if err == nil {
		c.String(http.StatusOK, "ok")
	} else {
		c.String(http.StatusInternalServerError, "mongo read unavailable")
		log.Println(err)
	}
}
