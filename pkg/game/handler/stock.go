package handler

import (
	"goserver/pkg/data"
)

func InitStock() {
	// games := config.GetGames2()
	// for _, v := range games {
	// 	stock := new(data.Stock)
	// 	stock.GetById(v.Id)
	// 	if stock.Id == "" {
	// 		stock.Id = v.Id
	// 		stock.Save()
	// 	}
	// }
	stocks := []data.Stock{
		//tp
		{Id: "1001", Type: 1, CashStock: 50000, Factor: 1, History: []data.StockHistory{}},
		{Id: "1002", Type: 1, CashStock: 200000, Factor: 1, History: []data.StockHistory{}},
		{Id: "1003", Type: 1, CashStock: 1000000, Factor: 1, History: []data.StockHistory{}},
		{Id: "1004", Type: 1, CashStock: 3000000, Factor: 1, History: []data.StockHistory{}},
		{Id: "1005", Type: 1, CashStock: 5000000, Factor: 1, History: []data.StockHistory{}},
		{Id: "1006", Type: 1, CashStock: 10000000, Factor: 1, History: []data.StockHistory{}},
		{Id: "1007", Type: 1, CashStock: 20000000, Factor: 1, History: []data.StockHistory{}},
		{Id: "1008", Type: 1, CashStock: 20000000, Factor: 1, History: []data.StockHistory{}},
		{Id: "1009", Type: 1, CashStock: 20000000, Factor: 1, History: []data.StockHistory{}},
		//tp2 sa版
		{Id: "1801", Type: 1, CashStock: 20000000, Factor: 1, History: []data.StockHistory{}},
		{Id: "1802", Type: 1, CashStock: 20000000, Factor: 1, History: []data.StockHistory{}},
		{Id: "1803", Type: 1, CashStock: 20000000, Factor: 1, History: []data.StockHistory{}},
		{Id: "1804", Type: 1, CashStock: 20000000, Factor: 1, History: []data.StockHistory{}},
		{Id: "1805", Type: 1, CashStock: 20000000, Factor: 1, History: []data.StockHistory{}},
		{Id: "1806", Type: 1, CashStock: 20000000, Factor: 1, History: []data.StockHistory{}},
		{Id: "1807", Type: 1, CashStock: 20000000, Factor: 1, History: []data.StockHistory{}},
		{Id: "1808", Type: 1, CashStock: 20000000, Factor: 1, History: []data.StockHistory{}},
		{Id: "1809", Type: 1, CashStock: 20000000, Factor: 1, History: []data.StockHistory{}},
		{Id: "1810", Type: 1, CashStock: 20000000, Factor: 1, History: []data.StockHistory{}},
		{Id: "1811", Type: 1, CashStock: 20000000, Factor: 1, History: []data.StockHistory{}},
		{Id: "1812", Type: 1, CashStock: 20000000, Factor: 1, History: []data.StockHistory{}},
		//joker
		{Id: "6001", Type: 1, CashStock: 50000, Factor: 1, History: []data.StockHistory{}},
		{Id: "6002", Type: 1, CashStock: 200000, Factor: 1, History: []data.StockHistory{}},
		{Id: "6003", Type: 1, CashStock: 1000000, Factor: 1, History: []data.StockHistory{}},
		{Id: "6004", Type: 1, CashStock: 3000000, Factor: 1, History: []data.StockHistory{}},
		{Id: "6005", Type: 1, CashStock: 5000000, Factor: 1, History: []data.StockHistory{}},
		{Id: "6006", Type: 1, CashStock: 10000000, Factor: 1, History: []data.StockHistory{}},
		{Id: "6007", Type: 1, CashStock: 20000000, Factor: 1, History: []data.StockHistory{}},
		{Id: "6008", Type: 1, CashStock: 20000000, Factor: 1, History: []data.StockHistory{}},
		{Id: "6009", Type: 1, CashStock: 20000000, Factor: 1, History: []data.StockHistory{}},
		//ak47
		{Id: "5001", Type: 1, CashStock: 50000, Factor: 1, History: []data.StockHistory{}},
		{Id: "5002", Type: 1, CashStock: 200000, Factor: 1, History: []data.StockHistory{}},
		{Id: "5003", Type: 1, CashStock: 1000000, Factor: 1, History: []data.StockHistory{}},
		{Id: "5004", Type: 1, CashStock: 3000000, Factor: 1, History: []data.StockHistory{}},
		{Id: "5005", Type: 1, CashStock: 5000000, Factor: 1, History: []data.StockHistory{}},
		{Id: "5006", Type: 1, CashStock: 10000000, Factor: 1, History: []data.StockHistory{}},
		{Id: "5007", Type: 1, CashStock: 20000000, Factor: 1, History: []data.StockHistory{}},
		{Id: "5008", Type: 1, CashStock: 20000000, Factor: 1, History: []data.StockHistory{}},
		{Id: "5009", Type: 1, CashStock: 20000000, Factor: 1, History: []data.StockHistory{}},
		//rm
		{Id: "2001", Type: 2, CashStock: 50000, Factor: 1, History: []data.StockHistory{}},
		{Id: "2002", Type: 2, CashStock: 100000, Factor: 1, History: []data.StockHistory{}},
		{Id: "2003", Type: 2, CashStock: 500000, Factor: 1, History: []data.StockHistory{}},
		{Id: "2004", Type: 2, CashStock: 1000000, Factor: 1, History: []data.StockHistory{}},
		{Id: "2005", Type: 2, CashStock: 2000000, Factor: 1, History: []data.StockHistory{}},
		{Id: "2006", Type: 2, CashStock: 5000000, Factor: 1, History: []data.StockHistory{}},
		{Id: "2007", Type: 2, CashStock: 10000000, Factor: 1, History: []data.StockHistory{}},
		{Id: "2101", Type: 2, CashStock: 50000, Factor: 1, History: []data.StockHistory{}},
		{Id: "2102", Type: 2, CashStock: 100000, Factor: 1, History: []data.StockHistory{}},
		{Id: "2103", Type: 2, CashStock: 500000, Factor: 1, History: []data.StockHistory{}},
		{Id: "2104", Type: 2, CashStock: 1000000, Factor: 1, History: []data.StockHistory{}},
		{Id: "2105", Type: 2, CashStock: 2000000, Factor: 1, History: []data.StockHistory{}},
		{Id: "2106", Type: 2, CashStock: 5000000, Factor: 1, History: []data.StockHistory{}},
		{Id: "2107", Type: 2, CashStock: 5000000, Factor: 1, History: []data.StockHistory{}},
		//rm双人
		{Id: "2801", Type: 2, CashStock: 0, Factor: 1, History: []data.StockHistory{}},
		{Id: "2802", Type: 2, CashStock: 0, Factor: 1, History: []data.StockHistory{}},
		{Id: "2803", Type: 2, CashStock: 0, Factor: 1, History: []data.StockHistory{}},
		{Id: "2804", Type: 2, CashStock: 0, Factor: 1, History: []data.StockHistory{}},
		{Id: "2805", Type: 2, CashStock: 0, Factor: 1, History: []data.StockHistory{}},
		{Id: "2806", Type: 2, CashStock: 0, Factor: 1, History: []data.StockHistory{}},
		{Id: "2807", Type: 2, CashStock: 0, Factor: 1, History: []data.StockHistory{}},
		{Id: "2808", Type: 2, CashStock: 0, Factor: 1, History: []data.StockHistory{}},
		{Id: "2809", Type: 2, CashStock: 0, Factor: 1, History: []data.StockHistory{}},
		{Id: "2810", Type: 2, CashStock: 0, Factor: 1, History: []data.StockHistory{}},
		{Id: "2901", Type: 2, CashStock: 0, Factor: 1, History: []data.StockHistory{}},
		{Id: "2902", Type: 2, CashStock: 0, Factor: 1, History: []data.StockHistory{}},
		{Id: "2903", Type: 2, CashStock: 0, Factor: 1, History: []data.StockHistory{}},
		{Id: "2904", Type: 2, CashStock: 0, Factor: 1, History: []data.StockHistory{}},
		{Id: "2905", Type: 2, CashStock: 0, Factor: 1, History: []data.StockHistory{}},
		{Id: "2906", Type: 2, CashStock: 0, Factor: 1, History: []data.StockHistory{}},
		{Id: "2907", Type: 2, CashStock: 0, Factor: 1, History: []data.StockHistory{}},
		//lhd
		{Id: "30011", Type: 3, CashStock: 50000, Factor: 1, History: []data.StockHistory{}},
		{Id: "30012", Type: 3, CashStock: 200000, Factor: 1, History: []data.StockHistory{}},
		{Id: "30013", Type: 3, CashStock: 500000, Factor: 1, History: []data.StockHistory{}},
		{Id: "30014", Type: 3, CashStock: 1000000, Factor: 1, History: []data.StockHistory{}},
		{Id: "30015", Type: 3, CashStock: 3000000, Factor: 1, History: []data.StockHistory{}},
		{Id: "30021", Type: 3, CashStock: 50000, Factor: 1, History: []data.StockHistory{}},
		{Id: "30022", Type: 3, CashStock: 300000, Factor: 1, History: []data.StockHistory{}},
		{Id: "30023", Type: 3, CashStock: 300000, Factor: 1, History: []data.StockHistory{}},
		{Id: "30033", Type: 3, CashStock: 0, Factor: 1, History: []data.StockHistory{}},
		//7up
		{Id: "40011", Type: 3, CashStock: 50000, Factor: 1, History: []data.StockHistory{}},
		{Id: "40012", Type: 3, CashStock: 200000, Factor: 1, History: []data.StockHistory{}},
		{Id: "40013", Type: 3, CashStock: 500000, Factor: 1, History: []data.StockHistory{}},
		{Id: "40014", Type: 3, CashStock: 1000000, Factor: 1, History: []data.StockHistory{}},
		{Id: "40015", Type: 3, CashStock: 3000000, Factor: 1, History: []data.StockHistory{}},
		{Id: "40021", Type: 3, CashStock: 50000, Factor: 1, History: []data.StockHistory{}},
		{Id: "40022", Type: 3, CashStock: 300000, Factor: 1, History: []data.StockHistory{}},
		{Id: "40023", Type: 3, CashStock: 300000, Factor: 1, History: []data.StockHistory{}},
		{Id: "40033", Type: 3, CashStock: 0, Factor: 1, History: []data.StockHistory{}},
		//crash
		{Id: "50011", Type: 3, CashStock: 50000, Factor: 1, History: []data.StockHistory{}},
		{Id: "50012", Type: 3, CashStock: 200000, Factor: 1, History: []data.StockHistory{}},
		{Id: "50013", Type: 3, CashStock: 500000, Factor: 1, History: []data.StockHistory{}},
		{Id: "50014", Type: 3, CashStock: 1000000, Factor: 1, History: []data.StockHistory{}},
		{Id: "50015", Type: 3, CashStock: 3000000, Factor: 1, History: []data.StockHistory{}},
		{Id: "50021", Type: 3, CashStock: 50000, Factor: 1, History: []data.StockHistory{}},
		{Id: "50022", Type: 3, CashStock: 300000, Factor: 1, History: []data.StockHistory{}},
		{Id: "50023", Type: 3, CashStock: 300000, Factor: 1, History: []data.StockHistory{}},
		{Id: "50033", Type: 3, CashStock: 0, Factor: 1, History: []data.StockHistory{}},
		//andarbahar
		{Id: "60011", Type: 3, CashStock: 50000, Factor: 1, History: []data.StockHistory{}},
		{Id: "60012", Type: 3, CashStock: 200000, Factor: 1, History: []data.StockHistory{}},
		{Id: "60013", Type: 3, CashStock: 500000, Factor: 1, History: []data.StockHistory{}},
		{Id: "60014", Type: 3, CashStock: 1000000, Factor: 1, History: []data.StockHistory{}},
		{Id: "60015", Type: 3, CashStock: 3000000, Factor: 1, History: []data.StockHistory{}},
		{Id: "60021", Type: 3, CashStock: 50000, Factor: 1, History: []data.StockHistory{}},
		{Id: "60022", Type: 3, CashStock: 300000, Factor: 1, History: []data.StockHistory{}},
		{Id: "60023", Type: 3, CashStock: 300000, Factor: 1, History: []data.StockHistory{}},
		{Id: "60033", Type: 3, CashStock: 0, Factor: 1, History: []data.StockHistory{}},
		//彩票
		{Id: "70011", Type: 3, CashStock: 50000, Factor: 1, History: []data.StockHistory{}},
		{Id: "70012", Type: 3, CashStock: 200000, Factor: 1, History: []data.StockHistory{}},
		{Id: "70013", Type: 3, CashStock: 500000, Factor: 1, History: []data.StockHistory{}},
		{Id: "70014", Type: 3, CashStock: 1000000, Factor: 1, History: []data.StockHistory{}},
		{Id: "70015", Type: 3, CashStock: 3000000, Factor: 1, History: []data.StockHistory{}},
		{Id: "70021", Type: 3, CashStock: 50000, Factor: 1, History: []data.StockHistory{}},
		{Id: "70022", Type: 3, CashStock: 300000, Factor: 1, History: []data.StockHistory{}},
		{Id: "70023", Type: 3, CashStock: 300000, Factor: 1, History: []data.StockHistory{}},
		{Id: "70033", Type: 3, CashStock: 0, Factor: 1, History: []data.StockHistory{}},
		//飞机
		{Id: "80011", Type: 3, CashStock: 50000, Factor: 1, History: []data.StockHistory{}},
		{Id: "80012", Type: 3, CashStock: 200000, Factor: 1, History: []data.StockHistory{}},
		{Id: "80013", Type: 3, CashStock: 500000, Factor: 1, History: []data.StockHistory{}},
		{Id: "80014", Type: 3, CashStock: 1000000, Factor: 1, History: []data.StockHistory{}},
		{Id: "80015", Type: 3, CashStock: 3000000, Factor: 1, History: []data.StockHistory{}},
		{Id: "80021", Type: 3, CashStock: 50000, Factor: 1, History: []data.StockHistory{}},
		{Id: "80022", Type: 3, CashStock: 300000, Factor: 1, History: []data.StockHistory{}},
		{Id: "80023", Type: 3, CashStock: 300000, Factor: 1, History: []data.StockHistory{}},
		{Id: "80033", Type: 3, CashStock: 0, Factor: 1, History: []data.StockHistory{}},
	}
	for _, v := range stocks {
		v.Init()
		// v.SaveStock()
		// v.SaveHistory()
	}
}
