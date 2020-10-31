package network

import (
"fmt"
"bytes"
"encoding/gob"
"log"
"math"
"github.com/karai/go-karai/transaction"
"strconv"
"sort"
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

func (s *Server) inMempool(tx_hash string) bool {
    for _, b := range s.Prtl.Mempool.Transactions {
        if b.Hash == tx_hash {
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

//sorts oracle map
func (s *Server) sortOracleDataMap(block_height string) map[string][]transaction.Request_Oracle_Data {
	contract_data_map := make(map[string][]transaction.Request_Oracle_Data)

	for _, oracle_data := range s.Prtl.Mempool.Transactions {
		if oracle_data.Height == block_height {
			contract_data_map[oracle_data.Epoc] = append(contract_data_map[oracle_data.Epoc], oracle_data)
		}
	}
	return contract_data_map
}

//gets mean, standard deviation on a array of floats
func stdevData(oracle_array []float32) (float32, float32) {
	var stdev, mean, sum float32
	for _, price := range oracle_array {
		sum += price		
	}

	mean = sum / float32(len(oracle_array))

	for _, val := range oracle_array {
		stdev += float32(math.Pow(float64(val - mean), 2))
	}

	stdev = float32(math.Sqrt(float64(stdev/ float32(len(oracle_array)))   ))


	return stdev, mean
}

//takes prices and makes them into float32
func stringsToFloats(oracle_array []transaction.Request_Oracle_Data) []float32 {
	var floats []float32
	for _, s := range oracle_array {
		if float, err := strconv.ParseFloat(s.Data, 32); err == nil {
			floats = append(floats, float32(float))
		}
	}
	return floats
}

//Checks if price is one deviation away 
func isOneDev(price, stdev, mean float32) bool {

	if price >= (stdev - mean) && price <= (stdev + mean) {
		return true
	}

	return false
}
//removes 
func remove(slice []transaction.Request_Oracle_Data, index int) []transaction.Request_Oracle_Data {
    return append(slice[:index], slice[index+1:]...)
}

func filterOracleDataMap(contract_map map[string][]transaction.Request_Oracle_Data) (map[string][]transaction.Request_Oracle_Data, map[string]float32) {
	contract_data_map := make(map[string][]transaction.Request_Oracle_Data)
	trusted_answer_data_map := make(map[string]float32)
	for _, oracle_array := range contract_map {
		floats := stringsToFloats(oracle_array)

		stdev, mean := stdevData(floats)

		for i, contract_data := range oracle_array {
			if !isOneDev(floats[i], stdev, mean) {
				contract_data_map[contract_data.Epoc] = append(contract_data_map[contract_data.Epoc], contract_data)
			} 
		}
		trusted_answer_data_map[oracle_array[0].Epoc] = calcMedian(floats)
	}
	return contract_data_map, trusted_answer_data_map
}

func calcMedian(floats []float32) float32 {

	float32Values := floats
	float32AsFloat64Values := make([]float64, len(floats))

	for i, val := range float32Values {
		float32AsFloat64Values[i] = float64(val)
	}

	sort.Float64s(float32AsFloat64Values)
	
	for i, val := range float32AsFloat64Values {
		float32Values[i] = float32(val)
	}
	
	mnum := len(float32Values) / 2

	if len(float32Values) % 2 == 0 {
		return float32Values[mnum]
	}

	return (float32Values[mnum-1] + float32Values[mnum]) / 2
}