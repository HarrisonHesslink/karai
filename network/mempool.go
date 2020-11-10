package network

import (
	"strconv"

	"github.com/harrisonhesslink/pythia/transaction"
	"github.com/harrisonhesslink/pythia/util"
)

func NewMemPool() *MemPool {
	m := new(MemPool)
	m.transactions_map = make(map[string]int)
	return m
}

//sorts oracle map
func (m *MemPool) SortOracleDataMap(block_height string) map[string][]transaction.OracleData {
	contract_data_map := make(map[string][]transaction.OracleData)

	for _, oracle_data := range m.transactions {
		if oracle_data.Height == block_height {
			contract_data_map[oracle_data.Contract] = append(contract_data_map[oracle_data.Contract], oracle_data)
		}
	}
	return contract_data_map
}

func (m *MemPool) addOracleData(tx transaction.OracleData) bool {

	if m.InMempool(tx.Hash) {
		return false
	}

	m.transactions = append(m.transactions, tx)
	m.transactions_map[tx.Hash] = (len(m.transactions) - 1)

	return true
}

func (m *MemPool) removeOracleData(tx_hash string) bool {

	remove(m.transactions, m.transactions_map[tx_hash])

	return true
}

func (m *MemPool) PrintMemPool() {
	for _, data := range m.transactions {

		util.Success_log_array("Hash: " + data.Hash[:8] + " For Height: " + data.Height)
	}
	util.Success_log("Printed: " + strconv.Itoa(len(m.transactions)) + " in oracle data mempool")
}

func (m *MemPool) Count() int {
	return len(m.transactions)
}

func FilterOracleDataMap(contract_map map[string][]transaction.OracleData) (map[string][]transaction.OracleData, map[string]float64) {
	contract_data_map := make(map[string][]transaction.OracleData)
	trusted_answer_data_map := make(map[string]float64)
	for _, oracle_array := range contract_map {
		floats := toFloatArray(oracle_array)

		stdev, mean := stdevData(floats)

		for i, contract_data := range oracle_array {
			if !isOneDev(floats[i], stdev, mean) {
				contract_data_map[contract_data.Contract] = append(contract_data_map[contract_data.Contract], contract_data)
			}
		}
		trusted_answer_data_map[oracle_array[0].Contract] = calcMedian(floats)
	}
	return contract_data_map, trusted_answer_data_map
}

func (m *MemPool) InMempool(tx_hash string) bool {
	for _, compare := range m.transactions {
		if compare.Hash == tx_hash {
			return true
		}
	}
	return false
}
