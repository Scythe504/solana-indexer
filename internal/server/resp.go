package server

import (
	"encoding/json"
	"log"
)

type resp struct {
	StatusCode    int         `json:"status_code"`
	RespStartTime int64       `json:"resp_time_start_ms"`
	RespEndTime   int64       `json:"resp_time_end_ms"`
	NetRespTime   int64       `json:"net_resp_time_ms"`
	Data          interface{} `json:"data"`
}

func (s resp) MarshalJson() ([]byte, error) {
	byteResp, err := json.Marshal(s)

	if err != nil {
		log.Println("Invalid resp object", err)
		return nil, err
	}

	return byteResp, nil
}
