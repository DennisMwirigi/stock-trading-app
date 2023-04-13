package main

import (
	"go.mongodb.org/mongo-driver/bson"
)

func mongo_read_cashbal(v []bson.D) []c_bal {
	var temp []c_bal
	for _, s := range v {
		var e c_bal
		for _, kv_pair := range s {
			// tempk := kv_pair.Key
			tempv := kv_pair.Value
			switch d := tempv.(type) {
			case float64:
				e.Cash_balance = d
			}
		}
		temp = append(temp, e)
	}
	return temp
}

func mongo_read_logs(v []bson.D) []logEntry {
	var temp []logEntry
	for _, s := range v {
		var e logEntry
		for _, kv_pair := range s {
			tempk := kv_pair.Key
			tempv := kv_pair.Value
			switch d := tempv.(type) {

			case string:
				{
					if tempk == "LogType" {
						e.LogType = d
					}
					if tempk == "Server" {
						e.Server = d
					}
					if tempk == "Command" {
						e.Command = d
					}
					if tempk == "Username" {
						e.Username = d
					}
					if tempk == "StockSymbol" {
						e.StockSymbol = d
					}
					if tempk == "Filename" {
						e.Filename = d
					}
					if tempk == "Cryptokey" {
						e.Cryptokey = d
					}
					if tempk == "Action" {
						e.Action = d
					}
					if tempk == "ErrorMessage" {
						e.ErrorMessage = d
					}
					if tempk == "DebugMessage" {
						e.DebugMessage = d
					}
				}
			case int64:
				{
					e.Timestamp = d
				}
			case int32:
				{
					if tempk == "TransactionNum" {
						e.TransactionNum = int(d)
					}
					if tempk == "QuoteServerTime" {
						e.QuoteServerTime = int(d)
					}
				}
			case float64:
				{
					if tempk == "Funds" {
						e.Funds = d
					}
					if tempk == "Price" {
						e.Price = d
					}
				}
			}
		}
		temp = append(temp, e)
	}
	// fmt.Println("Leaving mongo utils")
	return temp
}

