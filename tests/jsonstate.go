package tests

import (
	"errors"
	"fmt"
	"github.com/sudachen/playground/libeth/common"
	"github.com/sudachen/playground/libeth/state"
	"math/big"
	"strconv"
)

/*
   "log0_emptyMem" : {
       "env" : {
           "currentCoinbase" : "2adc25665018aa1fe0e6bc666dac8fc2697ff9ba",
           "currentDifficulty" : "0x0100",
           "currentGasLimit" : "0x0f4240",
           "currentNumber" : "0x00",
           "currentTimestamp" : "0x01",
           "previousHash" : "5e20a0453cecd065ea59c37ac63e079ee08998b6045136a8ce6635c7912ec0b6"
       },
       "logs" : [
           {
               "address" : "0f572e5295c57f15886f9b263e2f6d2d6c7b5ec6",
               "bloom" : "00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000008800000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000800000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
               "data" : "0x",
               "topics" : [
               ]
           }
       ],
       "out" : "0x",
       "post" : {
           "095e7baea6a6c7c4c2dfeb977efac326af552d87" : {
               "balance" : "0x0de0b6b3a7658689",
               "code" : "0x60006000600060006017730f572e5295c57f15886f9b263e2f6d2d6c7b5ec66103e8f1600055",
               "nonce" : "0x00",
               "storage" : {
                   "0x00" : "0x01"
               }
           },
       },
       "postStateRoot" : "3d3859bbd67f0e8744dce0f4ad3c8dd36645445adcdc8caa2c4863758bf9e5a2",
       "pre" : {
           "095e7baea6a6c7c4c2dfeb977efac326af552d87" : {
               "balance" : "0x0de0b6b3a7640000",
               "code" : "0x60006000600060006017730f572e5295c57f15886f9b263e2f6d2d6c7b5ec66103e8f1600055",
               "nonce" : "0x00",
               "storage" : {
               }
           },
       },
       "transaction" : {
           "data" : "",
           "gasLimit" : "0x033450",
           "gasPrice" : "0x01",
           "nonce" : "0x00",
           "secretKey" : "45a915e4d060149eb4365960e6a7a45f334393093061116b197e3240065ff2d8",
           "to" : "095e7baea6a6c7c4c2dfeb977efac326af552d87",
           "value" : "0x0186a0"
       }
   }
*/

func strToBigInt(m map[string]interface{}, index string) (*big.Int, error) {
	if v, ok := m[index]; !ok {
		return nil, fmt.Errorf("there is no %s field", index)
	} else {
		if s, ok := v.(string); ok {
			if ret, ok := new(big.Int).SetString(s, 0); ok {
				return ret, nil
			}
		}
		return nil, fmt.Errorf("field %s has malformed value %s", index, v)
	}
}

func strToUint64(m map[string]interface{}, index string) (uint64, error) {
	if v, ok := m[index]; !ok {
		return 0, fmt.Errorf("there is no %s field", index)
	} else {
		if s, ok := v.(string); ok {
			if ret, err := strconv.ParseUint(s, 0, 64); err == nil {
				return ret, nil
			}
		}
		return 0, fmt.Errorf("field %s has malformed value %s", index, v)
	}
}

func strToAddress(m map[string]interface{}, index string) (common.Address, error) {
	if v, ok := m[index]; !ok {
		return common.Address{}, fmt.Errorf("there is no %s field", index)
	} else {
		if s, ok := v.(string); ok {
			return common.HexToAddress(s), nil
		}
		return common.Address{}, fmt.Errorf("field %s has malformed value %s", index, v)
	}
}

func strToHash(m map[string]interface{}, index string) (common.Hash, error) {
	if v, ok := m[index]; !ok {
		return common.Hash{}, fmt.Errorf("there is no %s field", index)
	} else {
		if s, ok := v.(string); ok {
			return common.HexToHash(s), nil
		}
		return common.Hash{}, fmt.Errorf("field %s has malformed value %s", index, v)
	}
}

func strToAddressOpt(m map[string]interface{}, index string) (*common.Address, error) {
	if v, ok := m[index]; !ok {
		return nil, nil
	} else {
		if s, ok := v.(string); ok {
			if len(s) >= 20*2 {
				addr := common.HexToAddress(s)
				return &addr, nil
			}
			return nil, nil
		}
		return nil, fmt.Errorf("field %s has malformed value %s", index, v)
	}
}

func strToBytes(m map[string]interface{}, index string) ([]byte, error) {
	if v, ok := m[index]; !ok {
		return nil, fmt.Errorf("there is no %s field", index)
	} else {
		if s, ok := v.(string); ok {
			return common.FromHex(s), nil
		}
		return nil, fmt.Errorf("field %s has malformed value %s", index, v)
	}
}

func strToBytesOpt(m map[string]interface{}, index string) ([]byte, error) {
	if v, ok := m[index]; !ok {
		return make([]byte, 0), nil
	} else {
		if s, ok := v.(string); ok {
			return common.FromHex(s), nil
		}
		return nil, fmt.Errorf("field %s has malformed value %s", index, v)
	}
}

func setStorage(m map[string]interface{}, address common.Address, st common.MutableState) error {
	if v, ok := m["storage"]; !ok {
		return errors.New("there is no storage field")
	} else {
		if values, ok := v.(map[string]interface{}); ok {
			for key, val := range values {
				if s, success := val.(string); success {
					st.SetValue(address, common.HexToHash(key), common.HexToHash(s))
				} else {
					return fmt.Errorf("storage key %s has malformed value", key)
				}
			}
			return nil
		}
		return errors.New("field storage has malformed value")
	}
}

func getString(m map[string]interface{}, index string) (string, error) {
	if v, ok := m[index]; !ok {
		return "", fmt.Errorf("there is no %s field", index)
	} else {
		if s, ok := v.(string); ok {
			return s, nil
		} else {
			return "", fmt.Errorf("field %s has malformed value", index)
		}
	}
}

func getHashes(m map[string]interface{}, index string) ([]common.Hash, error) {
	return nil, nil
}

func GetTransactionLogs(m map[string]interface{}) ([]*common.Log, error) {
	ret := make([]*common.Log, 0, 3)
	if v, ok := m["logs"]; ok {
		if values, ok := v.([]interface{}); !ok {
			return nil, errors.New("field storage has malformed value")
		} else {
			for _, v := range values {
				if log, ok := v.(map[string]interface{}); ok {
					var address string
					var topics []common.Hash
					var err error
					if address, err = getString(log, "address"); err != nil {
						return nil, fmt.Errorf("bad log record: %s", err.Error())
					}
					if topics, err = getHashes(log, "topics"); err != nil {
						return nil, fmt.Errorf("bad log record: %s", err.Error())
					}
					if data, err := getString(log, "data"); err != nil {
						return nil, fmt.Errorf("bad log record: %s", err.Error())
					} else {
						ret = append(ret, &common.Log{
							Address: common.HexToAddress(address),
							Topics:  topics,
							Data:    common.FromHex(data)})
					}
				}
			}
		}
	}
	return ret, nil
}

func FillStateFrom(accounts map[string]interface{}, st common.MutableState) error {
	for a, acc := range accounts {
		if m, ok := acc.(map[string]interface{}); !ok {
			return fmt.Errorf("malformed accouns set")
		} else {
			address := common.HexToAddress(a)

			if balance, err := strToBigInt(m, "balance"); err != nil {
				return fmt.Errorf("%s : %s", a, err.Error())
			} else {
				st.SetBalance(address, balance)
			}

			if nonce, err := strToUint64(m, "nonce"); err != nil {
				return fmt.Errorf("%s : %s", a, err.Error())
			} else {
				st.SetNonce(address, nonce)
			}

			if code, err := strToBytesOpt(m, "code"); err != nil {
				return fmt.Errorf("%s : %s", a, err.Error())
			} else {
				st.SetCode(address, code)
			}

			if err := setStorage(m, address, st); err != nil {
				return fmt.Errorf("%s : %s", a, err.Error())
			}
		}
	}
	return nil
}

func GetPossibleForks(test map[string]interface{}) []string {
	var ret []string
	if m, ok := test["post"]; ok {
		if v, ok := m.(map[string]interface{}); ok {
			for s := range v {
				ret = append(ret, s)
			}
		}
	}
	return ret
}

func NewPreState(test map[string]interface{}) (common.State, error) {
	pre := state.NewMicroState(nil)
	var err error

	if m, ok := test["pre"]; ok {
		if v, ok := m.(map[string]interface{}); ok {
			if err = FillStateFrom(v, pre); err != nil {
				return nil, err
			}
		}
	}

	return pre.Freeze(), nil
}

func NewClassicPostState(test map[string]interface{}) (common.State, error) {
	post := state.NewMicroState(nil)
	var err error

	if m, ok := test["post"]; ok {
		if v, ok := m.(map[string]interface{}); ok {
			if err = FillStateFrom(v, post); err != nil {
				return nil, err
			}
		}
	}

	return post.Freeze(), nil
}

func GetTransaction(test map[string]interface{}) (*common.Transaction, error) {
	var err error

	if m, ok := test["transaction"]; ok {
		if v, ok := m.(map[string]interface{}); ok {
			tx := &common.Transaction{}
			if tx.Data, err = strToBytes(v, "data"); err != nil {
				return nil, err
			}
			if tx.GasLimit, err = strToBigInt(v, "gasLimit"); err != nil {
				return nil, err
			}
			if tx.GasPrice, err = strToBigInt(v, "gasPrice"); err != nil {
				return nil, err
			}
			if tx.Value, err = strToBigInt(v, "value"); err != nil {
				return nil, err
			}
			if tx.Nonce, err = strToUint64(v, "nonce"); err != nil {
				return nil, err
			}
			if tx.To, err = strToAddressOpt(v, "to"); err != nil {
				return nil, err
			}
			return tx, nil
		}
	}

	return nil, errors.New("transaction does not exist in test definition")
}

func GetSecretKey(test map[string]interface{}) ([]byte, error) {
	var err error
	var secretKey []byte
	if m, ok := test["transaction"]; ok {
		if v, ok := m.(map[string]interface{}); ok {
			if secretKey, err = strToBytes(v, "secretKey"); err != nil {
				return nil, err
			}
			return secretKey, nil
		}
	}
	return nil, errors.New("transaction secretKey does not exist in test definition")
}

func GetTransactionOut(test map[string]interface{}) ([]byte, error) {
	return strToBytes(test, "out")
}

func FillBlockInfo(test map[string]interface{}, blockInfo *common.BlockInfo) error {
	var err error
	if m, ok := test["env"]; ok {
		if env, ok := m.(map[string]interface{}); ok {
			if blockInfo.Coinbase, err = strToAddress(env, "currentCoinbase"); err != nil {
				return err
			}
			if blockInfo.Difficulty, err = strToBigInt(env, "currentGasLimit"); err != nil {
				return err
			}
			if blockInfo.GasLimit, err = strToBigInt(env, "currentGasLimit"); err != nil {
				return err
			}
			if blockInfo.Number, err = strToBigInt(env, "currentNumber"); err != nil {
				return err
			}
			if blockInfo.Time, err = strToBigInt(env, "currentTimestamp"); err != nil {
				return err
			}
			if blockInfo.ParentHash, err = strToHash(env, "previousHash"); err != nil {
				return err
			}
			return nil
		}
	}
	return errors.New("env does not exist in test definition or malformed")
}
