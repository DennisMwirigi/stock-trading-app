package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/urfave/cli"
)

// Cmd struct is a representation of an isolated command executed by a user
type Cmd struct {
	Command  string `json:"-"`
	Id       string `json:"id"`
	Stock    string `json:"stock"`
	Amount   string `json:"amount"`
	Filename string `json:"filename"`
}

func main() {
	app := cli.NewApp()
	app.Name = "DayTrading Inc. CLI"
	app.Usage = "Lets you execute user commands from a file containing a list of commands as well as execute individual user commands from the command line"

	app.Commands = []cli.Command{
		{
			Name:     "read",
			Aliases:  []string{"r"},
			HelpName: "read",
			Action: func(c *cli.Context) error {
				readFromFile(c)
				return nil
			},
			Usage:       `Reads file from specified path`,
			Description: `Read and parse user commands' file`,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "filelocation, fl",
					Usage: "full path of file containing commands",
				},
			},
		},
		{
			Name:     "execute",
			Aliases:  []string{"e"},
			HelpName: "execute",
			Action: func(c *cli.Context) error {

				command := strings.ToUpper(c.String("cmd"))
				id := c.String("id")
				stock := strings.ToUpper(c.String("stock"))
				amount := c.String("amount")
				filename := c.String("filename")

				cmd := Cmd{Command: command, Id: id, Stock: stock, Amount: amount, Filename: filename}

				executeCmd(cmd)
				return nil
			},
			Usage:       `Executes specified user command`,
			Description: `Executes the given command`,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "command, cmd",
					Usage: "user command",
				},
				cli.StringFlag{
					Name:  "id",
					Usage: "username",
				},
				cli.StringFlag{
					Name:  "stock",
					Usage: "stock's symbol",
				},
				cli.Float64Flag{
					Name:  "amount, amt",
					Usage: "amount in dollars",
				},
				cli.StringFlag{
					Name:  "filename, fn",
					Usage: "file to print out to",
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

// Function reads file containing commands then parses and executes them line by line
func readFromFile(c *cli.Context) error {
	fileLocation := c.String("filelocation")
	file, err := os.Open(fileLocation)

	if err != nil {
		log.Fatal(err)
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		cmd := parseLine(scanner.Text())
		executeCmd(cmd)
	}

	if err := scanner.Err(); err != nil {
		return (err)
	}

	return nil
}

// Function parses single line from file and returns struct containing command params
func parseLine(line string) Cmd {
	line_arr := strings.Split(line, " ")
	cmd_arr := strings.Split(line_arr[1], ",")
	command := cmd_arr[0]

	if command == "ADD" {
		return Cmd{Command: command, Id: cmd_arr[1], Amount: cmd_arr[2]}

	} else if command == "BUY" || command == "SELL" || command == "SET_BUY_AMOUNT" || command == "SET_BUY_TRIGGER" || command == "SET_SELL_AMOUNT" || command == "SET_SELL_TRIGGER" {
		return Cmd{Command: command, Id: cmd_arr[1], Stock: cmd_arr[2], Amount: cmd_arr[3]}

	} else if command == "QUOTE" || command == "CANCEL_SET_BUY" || command == "CANCEL_SET_SELL" {
		return Cmd{Command: command, Id: cmd_arr[1], Stock: cmd_arr[2]}

	} else if command == "COMMIT_BUY" || command == "COMMIT_SELL" || command == "CANCEL_BUY" || command == "CANCEL_SELL" || command == "DISPLAY_SUMMARY" {
		return Cmd{Command: command, Id: cmd_arr[1]}

	} else if command == "DUMPLOG" {
		if len(cmd_arr) == 2 {
			return Cmd{Command: command, Filename: cmd_arr[1]}
		} else {
			return Cmd{Command: command, Id: cmd_arr[1], Filename: cmd_arr[2]}
		}

	} else {
		fmt.Printf("Command received: %s, line: %s\n", command, line)
		panic("Unknown command received")

	}
}

// Function sends request to server to execute command given
func executeCmd(cmd Cmd) {
	var req *http.Request
	var err error

	reqUrlPrefix := "http://localhost:8080"

	parsedJson, err := json.Marshal(cmd)
	if err != nil {
		panic(err)
	}

	switch cmd.Command {
	case "ADD":
		req, err = http.NewRequest(http.MethodPut, reqUrlPrefix+"/users/addBal", bytes.NewBuffer(parsedJson))
	case "QUOTE":
		req, err = http.NewRequest(http.MethodGet, reqUrlPrefix+"/users/"+cmd.Id+"/quote/"+cmd.Stock, nil)
	case "BUY":
		req, err = http.NewRequest(http.MethodPost, reqUrlPrefix+"/users/buy", bytes.NewBuffer(parsedJson))
	case "COMMIT_BUY":
		req, err = http.NewRequest(http.MethodPost, reqUrlPrefix+"/users/buy/commit", bytes.NewBuffer(parsedJson))
	case "CANCEL_BUY":
		req, err = http.NewRequest(http.MethodDelete, reqUrlPrefix+"/users/"+cmd.Id+"buy/cancel", nil)
	case "SELL":
		req, err = http.NewRequest(http.MethodPost, reqUrlPrefix+"/users/sell", bytes.NewBuffer(parsedJson))
	case "COMMIT_SELL":
		req, err = http.NewRequest(http.MethodPost, reqUrlPrefix+"/users/sell/commit", bytes.NewBuffer(parsedJson))
	case "CANCEL_SELL":
		req, err = http.NewRequest(http.MethodDelete, reqUrlPrefix+"/users/"+cmd.Id+"/sell/cancel", nil)
	case "SET_BUY_AMOUNT":
		req, err = http.NewRequest(http.MethodPost, reqUrlPrefix+"/users/setbuy", bytes.NewBuffer(parsedJson))
	case "CANCEL_SET_BUY":
		req, err = http.NewRequest(http.MethodDelete, reqUrlPrefix+"/users/"+cmd.Id+"/setbuy/"+cmd.Stock+"/cancel", nil)
	case "SET_BUY_TRIGGER":
		req, err = http.NewRequest(http.MethodPost, reqUrlPrefix+"/users/setbuy/trigger", bytes.NewBuffer(parsedJson))
	case "SET_SELL_AMOUNT":
		req, err = http.NewRequest(http.MethodPost, reqUrlPrefix+"/users/setsell", bytes.NewBuffer(parsedJson))
	case "SET_SELL_TRIGGER":
		req, err = http.NewRequest(http.MethodPost, reqUrlPrefix+"/users/setsell/trigger", bytes.NewBuffer(parsedJson))
	case "CANCEL_SET_SELL":
		req, err = http.NewRequest(http.MethodDelete, reqUrlPrefix+"/users/"+cmd.Id+"/setsell/"+cmd.Stock+"/cancel", nil)
	case "DUMPLOG":
		req, err = http.NewRequest(http.MethodPost, reqUrlPrefix+"/dumplog", bytes.NewBuffer(parsedJson))
	case "DISPLAY_SUMMARY":
		req, err = http.NewRequest(http.MethodGet, reqUrlPrefix+"/displaysummary/"+cmd.Id, nil)
	}

	if err != nil {
		panic(err)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()
	// fmt.Printf("Got response code: %d\n", res.StatusCode)

	// resBody, err := ioutil.ReadAll(res.Body)
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Printf("Response body: %s\n", resBody)
}
