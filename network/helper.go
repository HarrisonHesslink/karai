package network

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
	"math"
	"sort"

	"github.com/harrisonhesslink/pythia/transaction"
)

func CmdToBytes(cmd string) []byte {
	var bytes [commandLength]byte

	for i, c := range cmd {
		bytes[i] = byte(c)
	}

	return bytes[:]
}

func BytesToCmd(bytes []byte) string {
	var cmd []byte

	for _, b := range bytes {
		if b != 0x0 {
			cmd = append(cmd, b)
		}
	}

	return fmt.Sprintf("%s", cmd)
}

func ExtractCmd(request []byte) []byte {
	return request[:commandLength]
}

func GobEncode(data interface{}) []byte {
	var buff bytes.Buffer

	enc := gob.NewEncoder(&buff)
	err := enc.Encode(data)
	if err != nil {
		log.Panic(err)
	}

	return buff.Bytes()
}

func NodeIsKnown(addr string) bool {
	for _, node := range KnownNodes {
		if node == addr {
			return true
		}
	}

	return false
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func containsValue(m map[string]string, v string) bool {
	for _, x := range m {
		if x == v {
			return true
		}
	}
	return false
}

//gets mean, standard deviation on a array of floats
func stdevData(oracle_array []float64) (float64, float64) {
	var stdev, mean, sum float64
	for _, price := range oracle_array {
		sum += price
	}

	mean = sum / float64(len(oracle_array))

	for _, val := range oracle_array {
		stdev += math.Pow(float64(val-mean), 2)
	}

	stdev = math.Sqrt(float64(stdev / float64(len(oracle_array))))

	return stdev, mean
}

//takes prices and makes them into float32
func toFloatArray(oracle_array []transaction.OracleData) []float64 {
	var floats []float64
	for _, s := range oracle_array {
		floats = append(floats, s.Price)
	}
	return floats
}

//Checks if price is one deviation away
func isOneDev(price, stdev, mean float64) bool {

	if price >= (stdev-mean) && price <= (stdev+mean) {
		return true
	}

	return false
}

//removes
func remove(slice []transaction.OracleData, index int) []transaction.OracleData {
	return append(slice[:index], slice[index+1:]...)
}

func calcMedian(floats []float64) float64 {
	if len(floats) > 1 {

		sort.Float64s(floats)

		mnum := len(floats) / 2

		if len(floats)%2 == 0 {
			return floats[mnum]
		}

		return (floats[mnum-1] + floats[mnum]) / 2
	}
	return 0
}
