package product

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/url"
	"time"

	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/errgroup"

	"github.com/sskotezhov/maturin/pkg/onec"
)

const (
	cacheKeyAll        = "catalog:all"
	cacheKeyCategories = "catalog:categories"
)

type Repository interface {
	GetAll(ctx context.Context) ([]Product, error)
	GetByID(ctx context.Context, id string) (*Product, error)
	GetCategories(ctx context.Context) ([]Category, error)
	RefreshCache(ctx context.Context) (productsCount, categoriesCount int, err error)
}

type oneCRepository struct {
	client   *onec.Client
	rdb      *redis.Client
	cacheTTL time.Duration
}

func NewRepository(client *onec.Client, rdb *redis.Client, cacheTTL time.Duration) Repository {
	return &oneCRepository{
		client:   client,
		rdb:      rdb,
		cacheTTL: cacheTTL,
	}
}

func (r *oneCRepository) GetAll(ctx context.Context) ([]Product, error) {
	cached, err := r.rdb.Get(ctx, cacheKeyAll).Bytes()
	if err == nil {
		var products []Product
		if err := json.Unmarshal(cached, &products); err == nil {
			slog.Info("catalog: cache hit", "count", len(products))
			return products, nil
		}
	}

	slog.Info("catalog: cache miss, fetching from 1C")
	products, err := r.fetchAndJoin(ctx)
	if err != nil {
		slog.Error("catalog: fetch failed", "err", err)
		return nil, err
	}

	slog.Info("catalog: fetch complete", "count", len(products))

	if data, err := json.Marshal(products); err == nil {
		r.rdb.Set(ctx, cacheKeyAll, data, r.cacheTTL)
	}

	return products, nil
}

func (r *oneCRepository) GetByID(ctx context.Context, id string) (*Product, error) {
	all, err := r.GetAll(ctx)
	if err != nil {
		return nil, err
	}
	for i := range all {
		if all[i].ID == id {
			return &all[i], nil
		}
	}
	return nil, ErrNotFound
}

func (r *oneCRepository) GetCategories(ctx context.Context) ([]Category, error) {
	cached, err := r.rdb.Get(ctx, cacheKeyCategories).Bytes()
	if err == nil {
		var cats []Category
		if err := json.Unmarshal(cached, &cats); err == nil {
			slog.Info("categories: cache hit", "count", len(cats))
			return cats, nil
		}
	}

	slog.Info("categories: cache miss, fetching from 1C")
	cats, err := r.fetchCategories(ctx)
	if err != nil {
		slog.Error("categories: fetch failed", "err", err)
		return nil, err
	}

	slog.Info("categories: fetch complete", "count", len(cats))

	if data, err := json.Marshal(cats); err == nil {
		r.rdb.Set(ctx, cacheKeyCategories, data, r.cacheTTL)
	}

	return cats, nil
}

func (r *oneCRepository) RefreshCache(ctx context.Context) (int, int, error) {
	if err := r.rdb.Del(ctx, cacheKeyAll, cacheKeyCategories).Err(); err != nil {
		slog.Warn("catalog: cache del failed", "err", err)
	}
	products, err := r.GetAll(ctx)
	if err != nil {
		return 0, 0, err
	}
	cats, err := r.GetCategories(ctx)
	if err != nil {
		return len(products), 0, err
	}
	return len(products), len(cats), nil
}

func (r *oneCRepository) fetchAndJoin(ctx context.Context) ([]Product, error) {
	var (
		rawProducts []oneCProduct
		rawPrices   []oneCPrice
		rawStocks   []oneCStockBalance
	)

	g, gctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		params := url.Values{}
		params.Set("$filter", "IsFolder eq false and DeletionMark eq false")
		params.Set("$select", "Ref_Key,Code,Description,НаименованиеПолное,Артикул,КатегорияНоменклатуры_Key,КатегорияНоменклатуры/Description,ТипНоменклатуры,ВидСтавкиНДС,ДатаИзменения")
		params.Set("$expand", "КатегорияНоменклатуры")
		items, err := r.client.Fetch(gctx, "Catalog_Номенклатура", params)
		if err != nil {
			slog.Error("1C: Catalog_Номенклатура failed", "err", err)
			return err
		}
		rawProducts = make([]oneCProduct, 0, len(items))
		for _, raw := range items {
			var p oneCProduct
			if err := json.Unmarshal(raw, &p); err == nil {
				rawProducts = append(rawProducts, p)
			}
		}
		slog.Info("1C: Catalog_Номенклатура fetched", "count", len(rawProducts))
		return nil
	})

	g.Go(func() error {
		params := url.Values{}
		params.Set("$select", "Номенклатура_Key,ВидЦен_Key,Period,Цена")
		items, err := r.client.Fetch(gctx, "InformationRegister_ЦеныНоменклатуры/SliceLast()", params)
		if err != nil {
			slog.Error("1C: InformationRegister_ЦеныНоменклатуры/SliceLast() failed", "err", err)
			return err
		}
		rawPrices = make([]oneCPrice, 0, len(items))
		for _, raw := range items {
			var p oneCPrice
			if err := json.Unmarshal(raw, &p); err == nil {
				rawPrices = append(rawPrices, p)
			}
		}
		slog.Info("1C: ЦеныНоменклатуры/SliceLast fetched", "count", len(rawPrices))
		return nil
	})

	g.Go(func() error {
		params := url.Values{}
		params.Set("$select", "Номенклатура_Key,КоличествоBalance")
		items, err := r.client.Fetch(gctx, "AccumulationRegister_Запасы/Balance", params)
		if err != nil {
			slog.Error("1C: AccumulationRegister_Запасы/Balance failed", "err", err)
			return err
		}
		rawStocks = make([]oneCStockBalance, 0, len(items))
		for _, raw := range items {
			var s oneCStockBalance
			if err := json.Unmarshal(raw, &s); err == nil {
				rawStocks = append(rawStocks, s)
			}
		}
		slog.Info("1C: Запасы/Balance fetched", "count", len(rawStocks))
		return nil
	})

	if err := g.Wait(); err != nil {
		return nil, err
	}

	result := join(rawProducts, rawPrices, rawStocks)
	slog.Info("catalog: join complete", "products", len(result))
	return result, nil
}

func (r *oneCRepository) fetchCategories(ctx context.Context) ([]Category, error) {
	params := url.Values{}
	params.Set("$select", "Ref_Key,Description")
	items, err := r.client.Fetch(ctx, "Catalog_КатегорииНоменклатуры", params)
	if err != nil {
		return nil, err
	}
	cats := make([]Category, 0, len(items))
	for _, raw := range items {
		var c oneCCategory
		if err := json.Unmarshal(raw, &c); err == nil {
			cats = append(cats, Category{ID: c.RefKey, Name: c.Description})
		}
	}
	return cats, nil
}

func join(products []oneCProduct, prices []oneCPrice, stocks []oneCStockBalance) []Product {
	latestPrice := make(map[string]oneCPrice, len(prices))
	for _, p := range prices {
		if p.PriceTypeKey != RetailPriceTypeKey {
			continue
		}
		if existing, ok := latestPrice[p.NomenclatureKey]; !ok || p.Period > existing.Period {
			latestPrice[p.NomenclatureKey] = p
		}
	}

	stockQty := make(map[string]float64, len(stocks))
	for _, s := range stocks {
		stockQty[s.NomenclatureKey] += s.QtyBalance
	}

	result := make([]Product, 0, len(products))
	for _, p := range products {
		categoryName := ""
		if p.CategoryExpanded != nil {
			categoryName = p.CategoryExpanded.Description
		}
		prod := Product{
			ID:           p.RefKey,
			Code:         p.Code,
			Name:         p.Description,
			FullName:     p.FullName,
			Article:      p.Article,
			CategoryKey:  p.CategoryKey,
			CategoryName: categoryName,
			Type:         p.Type,
			VAT:          p.VAT,
			UpdatedAt:    p.UpdatedAt,
		}

		if price, ok := latestPrice[p.RefKey]; ok && price.Price > 0 {
			v := price.Price
			prod.Price = &v
			prod.PriceDate = &price.Period
		}

		if p.Type == TypeService {
			prod.InStock = true
			prod.StockQty = nil
		} else {
			qty := int(max(stockQty[p.RefKey], 0))
			prod.StockQty = &qty
			prod.InStock = qty > 0
		}

		result = append(result, prod)
	}
	return result
}
