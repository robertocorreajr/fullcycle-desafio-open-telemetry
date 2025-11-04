package types

import (
	"encoding/json"
)

type WeatherResponse struct {
	TempC float64 `json:"temp_C"`
	TempF float64 `json:"temp_F"`
	TempK float64 `json:"temp_K"`
	City  string  `json:"city"`
}

type ViaCEPResult struct {
	Cep         string `json:"cep"`
	Logradouro  string `json:"logradouro"`
	Complemento string `json:"complemento"`
	Bairro      string `json:"bairro"`
	Localidade  string `json:"localidade"`
	UF          string `json:"uf"`
	IBGE        string `json:"ibge"`
	GIA         string `json:"gia"`
	DDD         string `json:"ddd"`
	SIAFI       string `json:"siafi"`
	Erro        bool   `json:"-"`
}

func (v *ViaCEPResult) UnmarshalJSON(data []byte) error {
	type Alias ViaCEPResult
	aux := struct {
		*Alias
		Erro interface{} `json:"erro"`
	}{
		Alias: (*Alias)(v),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	switch e := aux.Erro.(type) {
	case bool:
		v.Erro = e
	case string:
		v.Erro = e == "true"
	default:
		v.Erro = false
	}

	return nil
}

type WeatherAPIResponse struct {
	Current struct {
		TempC float64 `json:"temp_c"`
	} `json:"current"`
}
