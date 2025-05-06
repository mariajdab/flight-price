package helper

import (
	"fmt"
	"log"
	"strings"
)

func CityToIATACode(cityName string) string {
	cityToIATA := map[string]string{
		"paris":     "PAR",
		"madrid":    "MAD",
		"new york":  "JFK",
		"london":    "LHR",
		"tokyo":     "NRT",
		"berlin":    "BER",
		"rome":      "FCO",
		"moscow":    "SVO",
		"dubai":     "DXB",
		"barcelona": "BCN",
		"lisbon":    "LIS",
		"amsterdam": "AMS",
		"frankfurt": "FRA",
		"munich":    "MUC",
	}

	normalizedCity := strings.ToLower(strings.TrimSpace(cityName))

	code, exists := cityToIATA[normalizedCity]
	if !exists {
		log.Println(fmt.Errorf("no se encontró código IATA para la ciudad: %s", cityName))
		return ""
	}
	return code
}

func CityToGoogleCode(cityName string) string {
	cityToGoogleCode := map[string]string{
		"paris":     "CDG",
		"madrid":    "MAD",
		"new york":  "JFK",
		"london":    "LHR",
		"tokyo":     "HND",
		"berlin":    "BER",
		"rome":      "FCO",
		"moscow":    "SVO",
		"dubai":     "DXB",
		"barcelona": "BCN",
		"lisbon":    "LIS",
		"amsterdam": "AMS",
		"frankfurt": "FRA",
		"munich":    "MUC",
	}

	normalizedCity := strings.ToLower(strings.TrimSpace(cityName))

	code, exists := cityToGoogleCode[normalizedCity]
	if !exists {
		log.Println(fmt.Errorf("no se encontró código IATA para la ciudad: %s", cityName))
		return ""
	}
	return code
}

func CityToSkyCode(cityName string) string {
	cityToSkyCode := map[string]string{
		"paris":     "PARI",
		"madrid":    "MAD",
		"new york":  "NYCA",
		"london":    "LOND",
		"tokyo":     "TYOA",
		"berlin":    "BER",
		"rome":      "ROME",
		"moscow":    "MOSC",
		"dubai":     "DXBA",
		"barcelona": "BCN",
		"lisbon":    "LIS",
		"amsterdam": "AMS",
		"frankfurt": "FRAN",
		"munich":    "MUC",
	}

	normalizedCity := strings.ToLower(strings.TrimSpace(cityName))

	code, exists := cityToSkyCode[normalizedCity]
	if !exists {
		log.Println(fmt.Errorf("no se encontró código IATA para la ciudad: %s", cityName))
		return ""
	}
	return code
}
