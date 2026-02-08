// Copyright (C) 2026 Murilo Gomes Julio
// SPDX-License-Identifier: GPL-2.0-only

// Site: https://mugomes.github.io

package controls

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func ListarEstados() ([]string, error) {
	resp, err := http.Get("https://servicodados.ibge.gov.br/api/v1/localidades/estados")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("erro IBGE: %s", resp.Status)
	}

	var estados []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&estados); err != nil {
		return nil, err
	}

	var lstEstados []string
	for _, row := range estados {
		nome, okNome := row["nome"].(string)
		sigla, okSigla := row["sigla"].(string)
		if okNome && okSigla {
			lstEstados = append(lstEstados, fmt.Sprintf("%s - %s", sigla, nome))
		}
	}

	return lstEstados, nil
}

func ListarMunicipiosPorUF(uf string) ([]string, error) {
	endpoint := fmt.Sprintf(
		"https://servicodados.ibge.gov.br/api/v1/localidades/estados/%s/municipios",
		uf,
	)

	resp, err := http.Get(endpoint)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("erro IBGE: %s", resp.Status)
	}

	var raw []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, err
	}

	var municipios []string
	for _, item := range raw {
		nome, okNome := item["nome"].(string)
		if okNome {
			municipios = append(municipios, nome)
		}
	}

	return municipios, nil
}


func BuscarLocalizacaoPorNome(nome, uf string) (string, string, error) {
	listURL := fmt.Sprintf(
		"https://servicodados.ibge.gov.br/api/v1/localidades/estados/%s/municipios",
		uf,
	)

	resp, err := http.Get(listURL)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	var municipios []struct {
		ID   int    `json:"id"`
		Nome string `json:"nome"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&municipios); err != nil {
		return "", "", err
	}

	var municipioID int
	for _, m := range municipios {
		if strings.EqualFold(m.Nome, nome) {
			municipioID = m.ID
			break
		}
	}

	if municipioID == 0 {
		return "", "", fmt.Errorf("município %s/%s não encontrado", nome, uf)
	}

    coordsURL := fmt.Sprintf(
		"https://servicodados.ibge.gov.br/api/v3/malhas/municipios/%d/metadados",
		municipioID,
	)

	resp, err = http.Get(coordsURL)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	var coordsData []struct {
		Centroide struct {
			Latitude  float64 `json:"latitude"`
			Longitude float64 `json:"longitude"`
		} `json:"centroide"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&coordsData); err != nil {
		return "", "", err
	}

	if len(coordsData) == 0 {
		return "", "", fmt.Errorf("coordenadas não encontradas para o ID %d", municipioID)
	}

	lat := fmt.Sprintf("%f", coordsData[0].Centroide.Latitude)
	lon := fmt.Sprintf("%f", coordsData[0].Centroide.Longitude)

	return lat, lon, nil
}
