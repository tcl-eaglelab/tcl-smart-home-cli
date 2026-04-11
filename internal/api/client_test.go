package api

import (
	"testing"
)

func TestCodeMatches(t *testing.T) {
	tests := []struct {
		name    string
		code    interface{}
		targets []float64
		want    bool
	}{
		{"float64 match", float64(200), []float64{200}, true},
		{"float64 no match", float64(200), []float64{401}, false},
		{"float64 multi target match", float64(1), []float64{200, 1}, true},
		{"string match", "200", []float64{200}, true},
		{"string no match", "200", []float64{401}, false},
		{"string multi target match", "1", []float64{200, 1}, true},
		{"nil code", nil, []float64{200}, false},
		{"bool code", true, []float64{1}, false},
		{"empty targets", float64(200), []float64{}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := codeMatches(tt.code, tt.targets...)
			if got != tt.want {
				t.Errorf("codeMatches(%v, %v) = %v, want %v", tt.code, tt.targets, got, tt.want)
			}
		})
	}
}

func TestIsSuccess(t *testing.T) {
	tests := []struct {
		name string
		resp apiResponse
		want bool
	}{
		{"success true", apiResponse{Success: true}, true},
		{"code 200 float", apiResponse{Code: float64(200)}, true},
		{"code 1 float", apiResponse{Code: float64(1)}, true},
		{"code 200 string", apiResponse{Code: "200"}, true},
		{"code 1 string", apiResponse{Code: "1"}, true},
		{"code 401 float", apiResponse{Code: float64(401)}, false},
		{"code 0 float", apiResponse{Code: float64(0)}, false},
		{"empty response", apiResponse{}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isSuccess(tt.resp)
			if got != tt.want {
				t.Errorf("isSuccess(%+v) = %v, want %v", tt.resp, got, tt.want)
			}
		})
	}
}

func TestIsTokenError(t *testing.T) {
	tests := []struct {
		name string
		resp apiResponse
		want bool
	}{
		{"code 401", apiResponse{Code: float64(401)}, true},
		{"code 401 string", apiResponse{Code: "401"}, true},
		{"msg contains token", apiResponse{Msg: "token expired"}, true},
		{"message contains Token", apiResponse{Message: "Invalid Token"}, true},
		{"msg uppercase TOKEN", apiResponse{Msg: "TOKEN_INVALID"}, true},
		{"code 200 no token msg", apiResponse{Code: float64(200), Msg: "success"}, false},
		{"empty response", apiResponse{}, false},
		{"unrelated error msg", apiResponse{Code: float64(500), Msg: "server error"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isTokenError(tt.resp)
			if got != tt.want {
				t.Errorf("isTokenError(%+v) = %v, want %v", tt.resp, got, tt.want)
			}
		})
	}
}

func TestRandomMsgID(t *testing.T) {
	id1 := randomMsgID()
	id2 := randomMsgID()
	if id1 == "" {
		t.Error("randomMsgID() returned empty string")
	}
	if id1 == id2 {
		t.Errorf("randomMsgID() returned duplicate IDs: %s", id1)
	}
	if len(id1) != 32 {
		t.Errorf("randomMsgID() length = %d, want 32 hex chars", len(id1))
	}
}
