package product

const RetailPriceTypeKey = "c1e2f39c-fbe9-11ed-827c-fa163e38e500"

const (
	TypeGoods   = "Запас"
	TypeService = "Услуга"
)

type Product struct {
	ID           string   `json:"id"`
	Code         string   `json:"code"`
	Name         string   `json:"name"`
	FullName     string   `json:"full_name"`
	Article      string   `json:"article"`
	CategoryKey  string   `json:"category_key"`
	CategoryName string   `json:"category_name"`
	Type         string   `json:"type"`
	VAT          string   `json:"vat"`
	UpdatedAt    string   `json:"updated_at"`
	Price        *float64 `json:"price"`
	PriceDate    *string  `json:"price_date"`
	InStock      bool     `json:"in_stock"`
	StockQty     *int     `json:"stock_qty"`
}

type Category struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type oneCProduct struct {
	RefKey       string `json:"Ref_Key"`
	Code         string `json:"Code"`
	Description  string `json:"Description"`
	FullName     string `json:"НаименованиеПолное"`
	Article      string `json:"Артикул"`
	CategoryKey  string `json:"КатегорияНоменклатуры_Key"`
	Type         string `json:"ТипНоменклатуры"`
	VAT          string `json:"ВидСтавкиНДС"`
	UpdatedAt    string `json:"ДатаИзменения"`
	IsFolder     bool   `json:"IsFolder"`
	DeletionMark bool   `json:"DeletionMark"`
}

type oneCPrice struct {
	NomenclatureKey string  `json:"Номенклатура_Key"`
	PriceTypeKey    string  `json:"ВидЦен_Key"`
	Period          string  `json:"Period"`
	Price           float64 `json:"Цена"`
}

type oneCCategory struct {
	RefKey      string `json:"Ref_Key"`
	Description string `json:"Description"`
}

type oneCStockBalance struct {
	NomenclatureKey string  `json:"Номенклатура_Key"`
	QtyBalance      float64 `json:"КоличествоBalance"`
}
