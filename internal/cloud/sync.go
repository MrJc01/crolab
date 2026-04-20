// Copyright (c) 2026 Crolab Contributors. All rights reserved.
// Licensed under the Crolab Sustainable License (CSL).
// Contact: mrj.crom@gmail.com
package cloud

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

// VastAIBundle represents an offer from vast.ai
type VastAIBundle struct {
	ID           int     `json:"id"`
	GPUName      string  `json:"gpu_name"`
	GPUCount     int     `json:"num_gpus"`
	GPUVRAM      float64 `json:"gpu_ram"`
	DPHTotal     float64 `json:"dph_total"`
	MachineID    int     `json:"machine_id"`
	PublicIPAddr string  `json:"public_ipaddr"`
	Geolocation  string  `json:"geolocation"`
}

type VastAIResponse struct {
	Offers []VastAIBundle `json:"offers"`
}

// SyncVastAIOffers fetches real vast.ai offers and inserts them directly into our local Crolab.
func SyncVastAIOffers() (int, error) {
	// The query filters for rentable real machines with specific GPUs
	q := `{"rentable":{"eq":true},"gpu_name":{"in":["RTX 3090","RTX 4090","A100 SXM4","A100 PCIe","Tesla T4"]}}`
	url := fmt.Sprintf("https://console.vast.ai/api/v0/bundles/?q=%s", q)

	client := http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return 0, fmt.Errorf("vastAPI connection erro: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return 0, fmt.Errorf("vastAPI req rejected: HTTP %d", resp.StatusCode)
	}

	var parsed VastAIResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return 0, fmt.Errorf("vastAPI JSON error: %v", err)
	}

	// Flush existing machines where provider is vastai
	db.Exec("DELETE FROM machines WHERE provider = 'vastai'")

	totalCreated := 0
	for _, offer := range parsed.Offers {
		// Limit to typical useful machines to avoid bloating local DB (max 20)
		if totalCreated >= 20 {
			break
		}

		vramLabel := fmt.Sprintf("%.0fGB", offer.GPUVRAM/1024)
		// Arbitrage Business Rule: Crolab Price = Vast Price * 2
		priceHr := offer.DPHTotal * 2.0
		
		machine := DBMachine{
			ID:             fmt.Sprintf("vast-%d", offer.ID),
			Name:           fmt.Sprintf("%s Cluster (%s)", offer.GPUName, offer.Geolocation),
			GPU:            offer.GPUName,
			VRAM:           vramLabel,
			PriceHr:        priceHr,
			Status:         "available",
			Address:        offer.PublicIPAddr,
			Provider:       "vastai",
			ProviderCostHr: offer.DPHTotal,
		}

		if err := DBCreateMachine(machine); err == nil {
			totalCreated++
		}
	}

	// Auto Create generic plans if they don't exist based on tiers
	SyncAutoPlans()

	log.Printf("☁️ Sync: Puxadas %d novas GPUs Reais da Vast.AI", totalCreated)
	return totalCreated, nil
}

func SyncAutoPlans() {
	plans, _ := DBListPlans()
	if len(plans) == 0 {
		DBCreatePlan(DBPlan{
			ID:         "plan-entry",
			Name:       "Entry (T4 Nível Colab)",
			VRAM:       "16-24GB",
			Storage:    "100GB",
			PriceHr:    0.30,
			PriceMonth: 0,
			MaxUsers:   500,
			Active:     true,
		})
		DBCreatePlan(DBPlan{
			ID:         "plan-pro",
			Name:       "Pro (RTX 4090 Cloud)",
			VRAM:       "24GB",
			Storage:    "500GB",
			PriceHr:    0.80,
			PriceMonth: 0,
			MaxUsers:   100,
			Active:     true,
		})
		DBCreatePlan(DBPlan{
			ID:         "plan-enterprise",
			Name:       "Enterprise (A100)",
			VRAM:       "80GB",
			Storage:    "2TB",
			PriceHr:    3.00,
			PriceMonth: 0,
			MaxUsers:   50,
			Active:     true,
		})
	}
}
