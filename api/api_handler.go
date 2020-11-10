package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strconv"

	"github.com/Jeffail/gabs"

	"github.com/harrisonhesslink/pythia/contract"
	"github.com/harrisonhesslink/pythia/transaction"
)

//MakeRequest to contract api url and return data in string with bool if returned correctly
func MakeRequest(c contract.Contract) (map[string]string, bool) {

	//fmt.Print(string(s))

	data := make(map[string]string)

	for k, val := range c.APIURLS {
		resp, err := http.Get(val)

		if err != nil {
			return data, false
		}

		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)

		jsonParsed, err := gabs.ParseJSON(body)
		if err != nil {
			panic(err)
		}

		// Search JSON
		for _, v := range c.Data[k] {
			switch jsonParsed.Path(v).Data().(type) {
			case float64:
				data[v] = strconv.FormatFloat(jsonParsed.Path(v).Data().(float64), 'f', 6, 64)
				continue
			case string:
				data[v] = jsonParsed.Path(v).Data().(string)
				continue
			default:
				data[v] = ""
			}

		}
	}

	return data, true
}

func CoreSign(od transaction.OracleData) (string, string) {
	var sig string
	var hash string

	p, err := json.Marshal(od)
	if err != nil {
		fmt.Println(err)
		return "", ""
	}

	jsonObj := gabs.New()

	jsonObj.Set("2.0", "jsonrpc")
	jsonObj.Set("0", "id")
	jsonObj.Set("on_get_signature", "method")
	jsonObj.Set(string(p), "params", "message")

	resp, err := http.Post("http://127.0.0.1:9331/json_rpc", "application/json", bytes.NewBuffer([]byte(jsonObj.String())))

	if err != nil {
		log.Println(err)
		return "", ""
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	jsonParsed, err := gabs.ParseJSON(body)
	if err != nil {
		panic(err)
	}

	sig = jsonParsed.Path("result.signature").String()
	hash = jsonParsed.Path("result.hash").String()

	return sig, hash
}

func getTypeByRegexp(foo *json.RawMessage) (string, error) {
	re := regexp.MustCompile(`"type"\s*:\s*"([^"]*)"`)
	matches := re.FindAllSubmatch(*foo, -1)
	if len(matches) > 0 {
		return string(matches[0][1]), nil
	} else {
		return "", fmt.Errorf("cannot find type")
	}
}
