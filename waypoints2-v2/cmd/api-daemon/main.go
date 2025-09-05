package main

import (
	"blackbird-eu/waypoints2-v2/config"
	"blackbird-eu/waypoints2-v2/internal/rabbitmq"
	"blackbird-eu/waypoints2-v2/pkg/logger"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/google/uuid"
)

func init() {
	logger.Log.Out = os.Stdout
}

func loadInformation() rabbitmq.ScanInfo {
	CUSTOMER_API_KEY := "b1854969410a3d2fd5f66eac90406b991de7d6df66b07ca943957e9e4167c314"
	CUSTOMER_ID := "db53e4db-b173-4565-905c-7a976afc08dc"
	SCAN_ID := "d457aca4-c14c-4e3f-aa2a-9b767eadd438"
	TARGET := "https://example.com"
	TEMPLATE_IDS := ""
	// TARGET_ID := ""
	// VULNERABILITY_SCAN_ID := ""
	// SUBDOMAINX_SCAN_ID := ""
	// SPIDERX_SCAN_ID := ""
	// CHUNK_ID := ""
	// TEMPLATE_TAGS := "DEFAULT"
	// HEADERS := ""
	// DELAY := "0"
	// TIMEOUT := "15000"
	// VPN_CONNECTION_URI := "null"
	// CUSTOMER_NOTIFICATIONS_KEY := "567bcddc83f6e19d11748451b9e43b3c063c9183f1114b7bb210a22d3ee954c2"
	// CUSTOMER_CANARY_TOKEN := "ty982941"

	customerID, err := uuid.Parse(CUSTOMER_ID)
	if err != nil {
		logger.Log.Fatal("CUSTOMER_ID format is not UUID")
	}

	scanID, err := uuid.Parse(SCAN_ID)
	if err != nil {
		logger.Log.Fatal("SCAN_ID format is not UUID")
	}

	templateIds := []string{}

	for _, id := range strings.Split(TEMPLATE_IDS, ",") {
		templateId, err := uuid.Parse(id)
		if err != nil {
			fmt.Printf("[DEBUG:] Invalid templateId supplied: \"%s\" (%s)\n", id, err)
			continue
		}

		templateIds = append(templateIds, fmt.Sprintf("%s", templateId.String()))
	}

	scanInfo := rabbitmq.ScanInfo{
		CustomerAPIKey: CUSTOMER_API_KEY,
		CustomerID:     customerID,
		ScanID:         scanID,
		Target:         TARGET,
		TemplateIDs:    templateIds,
	}

	return scanInfo
}

func main() {
	// Load config
	config.LoadConfig()

	// Initialize RabbitMQ instance
	rmq, err := rabbitmq.NewRabbitMQ(&config.AppConfig.RabbitMQ)
	if err != nil {
		logger.Log.Errorf("Failed to initialize RabbitMQ: %v", err)
	}
	defer rmq.Close()

	scanInfo := loadInformation()

	scanInfoMsg, _ := json.Marshal(scanInfo)

	if err = rmq.Producer(string(scanInfoMsg)); err != nil {
		logger.Log.Fatal("RabbitMQ Producer has a problem with: ", err.Error())
	}
}
