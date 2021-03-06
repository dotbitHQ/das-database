package api_code

type ApiCode = int

const (
	ApiCodeSuccess        ApiCode = 0
	ApiCodeError500       ApiCode = 500
	ApiCodeParamsInvalid  ApiCode = 10000
	ApiCodeMethodNotExist ApiCode = 10001
	ApiCodeDbError        ApiCode = 10002
	ApiCodeCacheError     ApiCode = 10003
	ApiCodeBlockError     ApiCode = 10005

	ApiCodeSystemUpgrade ApiCode = 30019
)

type ApiResp struct {
	ErrNo  ApiCode     `json:"err_no"`
	ErrMsg string      `json:"err_msg"`
	Data   interface{} `json:"data,omitempty"`
}

func ApiRespOK() ApiResp {
	return ApiResp{
		ErrNo:  ApiCodeSuccess,
		ErrMsg: "",
	}
}

func ApiRespOKData(data interface{}) ApiResp {
	return ApiResp{
		ErrNo:  ApiCodeSuccess,
		ErrMsg: "",
		Data:   data,
	}
}

func ApiRespErr(apiCode ApiCode, apiMsg string) ApiResp {
	return ApiResp{
		ErrNo:  apiCode,
		ErrMsg: apiMsg,
		Data:   nil,
	}
}
