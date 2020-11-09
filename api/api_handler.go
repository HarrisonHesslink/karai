package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"

	"github.com/Jeffail/gabs"

	"github.com/harrisonhesslink/pythia/contract"
)

//Make a Request to contract api url and return data in string with bool if returned correctly
func MakeRequest(c contract.Contract) (map[string]string, bool) {

	s, _ := json.MarshalIndent(c, "", "\t")
	fmt.Print(string(s))

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
			fmt.Println("Get value of "+v+" from "+k+" :\t", jsonParsed.Path(v).Data())
			data[v] = jsonParsed.Path(v).Data().(string)
		}
	}

	return data, true
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
