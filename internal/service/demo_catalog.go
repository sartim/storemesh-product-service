package service

import (
	"context"

	productv1 "github.com/sartim/storemesh-product-service/gen/storemesh/product/v1"
)

// NewDemoCatalog creates the local-only catalog used when the service is run
// without DATABASE_URL. Production and persistent environments remain empty
// until products are explicitly imported through the API.
func NewDemoCatalog() *Catalog {
	catalog := NewCatalog()
	for _, fixture := range demoProducts {
		if _, err := catalog.CreateProduct(context.Background(), &productv1.CreateProductRequest{Product: &productv1.Product{
			Sku: fixture.sku, Name: fixture.name, Description: fixture.description,
			PriceMinor: fixture.priceMinor, Currency: "USD",
		}}); err != nil {
			panic("seed local demo catalog: " + err.Error())
		}
	}
	return catalog
}

type demoProduct struct {
	name, sku, description string
	priceMinor             int64
}

var demoProducts = []demoProduct{
	{"Halo Desk Lamp", "SM-LAMP-001", "Warm ambient light for focused evenings.", 8900},
	{"Cove Ceramic Mug", "SM-MUG-002", "A quiet morning ritual, made tactile.", 2400},
	{"Field Notes Set", "SM-NOTE-003", "Three soft-cover notebooks for ideas in motion.", 1800},
	{"Arc Wireless Charger", "SM-CHRG-004", "A clean charging spot for your everyday carry.", 4200},
	{"Dune Desk Tray", "SM-TRAY-005", "Keep the small essentials beautifully together.", 3600},
	{"Canvas Market Tote", "SM-TOTE-006", "A durable carry-all for the daily route.", 2900},
	{"Lumen Portable Speaker", "SM-SPKR-007", "Room-filling sound in a compact silhouette.", 12900},
	{"Still Water Bottle", "SM-BOTT-008", "Double-wall steel, cool from first sip to last.", 3200},
	{"Orbit Mechanical Keyboard", "SM-KEYB-009", "A satisfying, considered typing experience.", 14900},
	{"Cloud Wool Throw", "SM-THRW-010", "A soft layer for slow Sunday afternoons.", 9800},
	{"Moss Plant Pot", "SM-POT-011", "A little green energy for the workspace.", 2700},
	{"Ridge Headphones", "SM-HEAD-012", "Focused listening with all-day comfort.", 18900},
	{"Slate Cable Kit", "SM-CABL-013", "The tidy answer to desk-side cable clutter.", 2100},
	{"Sunday Coffee Beans", "SM-COFF-014", "Bright, balanced beans for a better first cup.", 1600},
	{"Studio Backpack", "SM-BACK-015", "A calm, capable home for your daily essentials.", 11900},
	{"Paperweight Stone", "SM-STON-016", "A small grounded object for a busy desk.", 1400},
	{"Ember Table Clock", "SM-CLOK-017", "A quiet visual anchor for focused mornings.", 6400},
	{"Vale Linen Apron", "SM-APRN-018", "A sturdy layer for cooking, making, and hosting.", 5200},
	{"Hearth Candle", "SM-CNDL-019", "Warm cedar and amber for a softer room.", 3800},
	{"North Ceramic Vase", "SM-VASE-020", "A simple vessel for a single branch or bloom.", 4600},
	{"Drift Reading Light", "SM-LITE-021", "A portable pool of light for late chapters.", 7600},
	{"Woven Storage Basket", "SM-BASK-022", "Open storage with a calm, natural texture.", 6900},
	{"Cedar Laptop Stand", "SM-STND-023", "Raise your screen and make space to breathe.", 8200},
	{"Rain Travel Umbrella", "SM-UMBR-024", "Compact coverage for unexpected weather.", 4100},
	{"Morrow Tea Infuser", "SM-TEA-025", "A considered steep for leaves and quiet pauses.", 2200},
	{"Pebble Bluetooth Tracker", "SM-TRKR-026", "Keep everyday essentials close and findable.", 5500},
	{"Aster Cotton Sheets", "SM-SHTS-027", "Crisp, breathable comfort for better rest.", 15900},
	{"Common Leather Wallet", "SM-WLET-028", "A slim home for the cards you actually carry.", 7200},
	{"Tide Picnic Blanket", "SM-PICN-029", "A soft, durable base for outside hours.", 11200},
	{"Mono USB-C Hub", "SM-HUB-030", "One compact connection point for a busy setup.", 6800},
	{"Juniper Hand Soap", "SM-SOAP-031", "A fresh botanical wash for daily rituals.", 2600},
	{"Loop Key Organizer", "SM-KEYR-032", "A quieter, cleaner way to carry your keys.", 3300},
}
