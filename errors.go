package stratum

import "fmt"

type StratumErrorCode int

const (
	STRATUM_ERROR_UNKNOWN                StratumErrorCode = -1
	STRATUM_ERROR_SERVICE                StratumErrorCode = -2
	STRATUM_ERROR_METHOD                 StratumErrorCode = -3
	STRATUM_ERROR_FEE_REQUIRED           StratumErrorCode = -10
	STRATUM_ERROR_SIGNATURE_REQUIRED     StratumErrorCode = -20
	STRATUM_ERROR_SIGNATURE_UNAVAILABLE  StratumErrorCode = -21
	STRATUM_ERROR_UNKNOWN_SIGNATURE_TYPE StratumErrorCode = -22
	STRATUM_ERROR_BAD_SIGNATURE          StratumErrorCode = -23
)

type StratumError struct {
	Code      StratumErrorCode `json:"code"`
	Message   string           `json:"message"`
	Traceback interface{}      `json:"traceback"`
}

func (se *StratumError) Error() string {
	return fmt.Sprintf("code=%v msg=%v traceback=%v", se.Code, se.Message, se.Traceback)
}
