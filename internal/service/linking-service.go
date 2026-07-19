package service

import (
	"context"
	"sort"
	"strings"
	"sync"

	"grocerics-backend/internal/domain"
	"grocerics-backend/internal/errs"
	"grocerics-backend/internal/integration/quickcommerce"
	"grocerics-backend/internal/repository"
	"grocerics-backend/internal/util"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type LinkingService struct {
	qc             quickcommerce.Client
	platforms      *repository.PlatformRepository
	cities         *repository.CityRepository
	links          *repository.ProductPlatformLinkRepository
	platformPrices *repository.PlatformPriceRepository
	variants       *repository.ProductVariantRepository
	pricing        *PricingService
}

func NewLinkingService(db *gorm.DB, qc quickcommerce.Client, pricing *PricingService) *LinkingService {
	return &LinkingService{
		qc:             qc,
		platforms:      repository.NewPlatformRepository(db),
		cities:         repository.NewCityRepository(db),
		links:          repository.NewProductPlatformLinkRepository(db),
		platformPrices: repository.NewPlatformPriceRepository(db),
		variants:       repository.NewProductVariantRepository(db),
		pricing:        pricing,
	}
}

type LinkSeed struct {
	PricePaise int64
	MRPPaise   int64
	Available  bool
	Inventory  *int
	DeepLink   string
	ImageURL   string
}

func firstNonEmpty(ss []string) string {
	for _, s := range ss {
		if s != "" {
			return s
		}
	}
	return ""
}

type Candidate struct {
	QCItemID    string `json:"qc_item_id"`
	Name        string `json:"name"`
	Brand       string `json:"brand"`
	Quantity    string `json:"quantity"`
	Multipack   int    `json:"multipack"`
	PricePaise  int64  `json:"price_paise"`
	MRPPaise    int64  `json:"mrp_paise"`
	Available   bool   `json:"available"`
	Inventory   *int   `json:"inventory,omitempty"`
	StockLabel  string `json:"stock_label,omitempty"`
	DeepLink    string `json:"deeplink,omitempty"`
	ImageURL    string `json:"image_url,omitempty"`
	IsCombo     bool   `json:"is_combo"`
	IsMultipack bool   `json:"is_multipack"`
	OutOfStock  bool   `json:"out_of_stock"`
}

func (s *LinkingService) qcLocation(cityID string) (quickcommerce.Location, error) {
	c, err := s.cities.FindByID(cityID)
	if err != nil {
		return quickcommerce.Location{}, err
	}
	if c == nil {
		return quickcommerce.Location{}, errs.NotFound("CITY_NOT_FOUND", "city not found")
	}
	if c.Lat == nil || c.Lng == nil || c.DefaultPincode == nil || *c.DefaultPincode == "" {
		return quickcommerce.Location{}, errs.BadRequest("CITY_NO_LOCATION", "city has no default lat/lng/pincode configured")
	}
	return quickcommerce.Location{Lat: *c.Lat, Lon: *c.Lng, Pincode: *c.DefaultPincode}, nil
}

func toCandidate(p quickcommerce.Product) Candidate {
	return Candidate{
		QCItemID:    p.ID,
		Name:        p.Name,
		Brand:       p.Brand,
		Quantity:    p.Quantity,
		Multipack:   p.Multipack,
		PricePaise:  p.PricePaise,
		MRPPaise:    p.MRPPaise,
		Available:   p.Available,
		Inventory:   p.Inventory,
		StockLabel:  p.StockLabel,
		DeepLink:    p.DeepLink,
		ImageURL:    firstNonEmpty(p.Images),
		IsCombo:     strings.Contains(strings.ToLower(p.Quantity), "combo"),
		IsMultipack: p.Multipack > 1,
		OutOfStock:  !p.Available,
	}
}

func (s *LinkingService) SearchCandidates(query, cityID string, platformCodes []string) (map[string][]Candidate, int, error) {
	loc, err := s.qcLocation(cityID)
	if err != nil {
		return nil, 0, err
	}
	if len(platformCodes) == 0 {
		all, err := s.platforms.ListSearchable()
		if err != nil {
			return nil, 0, err
		}
		for _, p := range all {
			platformCodes = append(platformCodes, p.Code)
		}
	}
	results := make(map[string][]Candidate, len(platformCodes))
	codeByQC := make(map[string]string, len(platformCodes))
	qcNames := make([]string, 0, len(platformCodes))
	for _, code := range platformCodes {
		results[code] = []Candidate{}
		pl, err := s.platforms.FindByCode(code)
		if err != nil {
			return nil, 0, err
		}
		if pl == nil || pl.QCName == nil || *pl.QCName == "" {
			continue // unknown or unmapped platform -> empty column
		}
		qcNames = append(qcNames, *pl.QCName)
		codeByQC[*pl.QCName] = code
	}
	if len(qcNames) == 0 {
		return results, 0, nil
	}
	gs, err := s.qc.GroupSearch(context.Background(), query, loc, qcNames)
	if err != nil {
		return nil, 0, err
	}
	for qcName, products := range gs.ByPlatform {
		code, ok := codeByQC[qcName]
		if !ok {
			continue
		}
		cands := make([]Candidate, 0, len(products))
		for _, pr := range products {
			cands = append(cands, toCandidate(pr))
		}
		sortCandidates(cands)
		results[code] = cands
	}
	return results, gs.CreditsRemaining, nil
}

func sortCandidates(cs []Candidate) {
	rank := func(c Candidate) int {
		r := 0
		if c.OutOfStock {
			r += 4
		}
		if c.IsCombo {
			r += 2
		}
		if c.IsMultipack {
			r++
		}
		return r
	}
	sort.SliceStable(cs, func(i, j int) bool { return rank(cs[i]) < rank(cs[j]) })
}

type BatchLinkItem struct {
	VariantID    string
	PlatformCode string
	QCItemID     string
	DeepLink     string
	Seed         *LinkSeed
}

type BatchLinkFailure struct {
	VariantID    string `json:"variant_id"`
	PlatformCode string `json:"platform_code"`
	Error        string `json:"error"`
}

type BatchLinkResult struct {
	Mapped int                `json:"mapped"`
	Failed []BatchLinkFailure `json:"failed,omitempty"`
}

func (s *LinkingService) Credits() (int, error) {
	c, err := s.qc.Credits(context.Background())
	if err != nil {
		return 0, errs.Internal("QC_CREDITS_FAILED", err)
	}
	return c.Remaining, nil
}

func (s *LinkingService) resolvePlatform(platformCode string) (*domain.Platform, error) {
	pl, err := s.platforms.FindByCode(platformCode)
	if err != nil {
		return nil, err
	}
	if pl == nil {
		return nil, errs.NotFound("PLATFORM_NOT_FOUND", "platform not found")
	}
	if pl.QCName == nil || *pl.QCName == "" {
		return nil, errs.BadRequest("PLATFORM_NO_QC", "platform has no QuickCommerce mapping")
	}
	return pl, nil
}

func (s *LinkingService) ConfirmLink(variantID, platformCode, cityID, qcItemID, deepLink string, seed *LinkSeed) error {
	pl, err := s.resolvePlatform(platformCode)
	if err != nil {
		return err
	}
	if err := s.linkOne(variantID, cityID, qcItemID, deepLink, pl, seed); err != nil {
		return err
	}
	s.startFanOut(cityID, []fanOutTarget{{variantID: variantID, platform: pl, qcItemID: qcItemID}})
	return nil
}

func (s *LinkingService) ConfirmLinks(cityID string, items []BatchLinkItem) (BatchLinkResult, error) {
	var res BatchLinkResult
	if err := s.requireCity(cityID); err != nil {
		return res, err
	}
	targets := make([]fanOutTarget, 0, len(items))
	for _, it := range items {
		pl, err := s.resolvePlatform(it.PlatformCode)
		if err == nil {
			err = s.linkOne(it.VariantID, cityID, it.QCItemID, it.DeepLink, pl, it.Seed)
		}
		if err != nil {
			res.Failed = append(res.Failed, BatchLinkFailure{
				VariantID: it.VariantID, PlatformCode: it.PlatformCode, Error: err.Error(),
			})
			continue
		}
		res.Mapped++
		targets = append(targets, fanOutTarget{variantID: it.VariantID, platform: pl, qcItemID: it.QCItemID})
	}
	s.startFanOut(cityID, targets)
	return res, nil
}

func (s *LinkingService) linkOne(variantID, cityID, qcItemID, deepLink string, pl *domain.Platform, seed *LinkSeed) error {
	price, itemDeepLink, err := s.seedPrice(variantID, pl, cityID, qcItemID, seed)
	if err != nil {
		return err
	}
	if deepLink == "" {
		deepLink = itemDeepLink
	}
	sku := qcItemID
	var image string
	if seed != nil {
		image = seed.ImageURL
	}
	if _, err := s.links.Upsert(&domain.ProductPlatformLink{
		VariantID:   variantID,
		PlatformID:  pl.ID,
		PlatformSKU: &sku,
		DeepLink:    util.PtrIfSet(deepLink),
		ImageURL:    util.PtrIfSet(image),
	}); err != nil {
		return err
	}
	if err := s.platformPrices.Upsert(price); err != nil {
		return err
	}
	return s.pricing.RecomputeVariantSummary(variantID, cityID)
}

func (s *LinkingService) seedPrice(variantID string, pl *domain.Platform, cityID, qcItemID string, seed *LinkSeed) (*domain.PlatformPrice, string, error) {
	if seed != nil {
		if err := s.requireCity(cityID); err != nil {
			return nil, "", err
		}
		item := &quickcommerce.ItemDetail{
			ID: qcItemID, PricePaise: seed.PricePaise, MRPPaise: seed.MRPPaise,
			Available: seed.Available, Inventory: seed.Inventory, DeepLink: seed.DeepLink,
		}
		return itemToPrice(variantID, pl.ID, cityID, item), item.DeepLink, nil
	}
	loc, err := s.qcLocation(cityID)
	if err != nil {
		return nil, "", err
	}
	item, err := s.qc.GetItem(context.Background(), qcItemID, *pl.QCName, loc)
	if err != nil {
		return nil, "", errs.Internal("QC_GET_ITEM_FAILED", err)
	}
	return itemToPrice(variantID, pl.ID, cityID, item), item.DeepLink, nil
}

func (s *LinkingService) requireCity(cityID string) error {
	c, err := s.cities.FindByID(cityID)
	if err != nil {
		return err
	}
	if c == nil {
		return errs.NotFound("CITY_NOT_FOUND", "city not found")
	}
	return nil
}

const fanOutConcurrency = 4

type fanOutTarget struct {
	variantID string
	platform  *domain.Platform
	qcItemID  string
}

func (s *LinkingService) startFanOut(seededCityID string, targets []fanOutTarget) {
	if len(targets) == 0 {
		return
	}
	go func() {
		defer func() {
			if r := recover(); r != nil {
				zap.S().Errorw("linking: fan-out panicked", "panic", r)
			}
		}()
		s.fanOut(seededCityID, targets)
	}()
}

func (s *LinkingService) fanOut(seededCityID string, targets []fanOutTarget) {
	cities, err := s.cities.ListEnabled()
	if err != nil {
		zap.S().Warnw("linking: fan-out could not list cities", "error", err)
		return
	}
	type job struct {
		city domain.City
		loc  quickcommerce.Location
		t    fanOutTarget
	}
	jobs := make([]job, 0, len(cities)*len(targets))
	for _, city := range cities {
		if city.ID == seededCityID {
			continue
		}
		loc, err := s.qcLocation(city.ID)
		if err != nil {
			zap.S().Warnw("linking: fan-out skipping city", "city", city.Slug, "error", err)
			continue
		}
		for _, t := range targets {
			jobs = append(jobs, job{city: city, loc: loc, t: t})
		}
	}
	if len(jobs) == 0 {
		return
	}

	var wg sync.WaitGroup
	sem := make(chan struct{}, fanOutConcurrency)
	for _, j := range jobs {
		wg.Add(1)
		go func(j job) {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					zap.S().Errorw("linking: fan-out job panicked", "panic", r)
				}
			}()
			sem <- struct{}{}
			defer func() { <-sem }()

			item, err := s.qc.GetItem(context.Background(), j.t.qcItemID, *j.t.platform.QCName, j.loc)
			if err != nil {
				zap.S().Warnw("linking: fan-out GetItem failed",
					"city", j.city.Slug, "item", j.t.qcItemID, "error", err)
				return
			}
			if err := s.platformPrices.Upsert(itemToPrice(j.t.variantID, j.t.platform.ID, j.city.ID, item)); err != nil {
				zap.S().Warnw("linking: fan-out price upsert failed", "city", j.city.Slug, "error", err)
			}
		}(j)
	}
	wg.Wait()

	done := make(map[string]bool, len(jobs))
	for _, j := range jobs {
		key := j.t.variantID + "\x00" + j.city.ID
		if done[key] {
			continue
		}
		done[key] = true
		if err := s.pricing.RecomputeVariantSummary(j.t.variantID, j.city.ID); err != nil {
			zap.S().Warnw("linking: fan-out recompute failed",
				"variant", j.t.variantID, "city", j.city.Slug, "error", err)
		}
	}
}

type RefreshResult struct {
	Refreshed        int `json:"refreshed"`
	Failed           int `json:"failed"`
	CreditsRemaining int `json:"credits_remaining"`
}

func (s *LinkingService) RefreshVariant(variantID, cityID string) (RefreshResult, error) {
	var res RefreshResult
	loc, err := s.qcLocation(cityID)
	if err != nil {
		return res, err
	}
	links, err := s.links.ListByVariant(variantID)
	if err != nil {
		return res, err
	}
	platformIDs := make([]string, 0, len(links))
	for _, l := range links {
		platformIDs = append(platformIDs, l.PlatformID)
	}
	plats, err := s.platforms.FindByIDs(platformIDs)
	if err != nil {
		return res, err
	}
	ctx := context.Background()
	for _, l := range links {
		pl, ok := plats[l.PlatformID]
		if l.PlatformSKU == nil || *l.PlatformSKU == "" || !ok || pl.QCName == nil || *pl.QCName == "" {
			res.Failed++
			continue
		}
		item, err := s.qc.GetItem(ctx, *l.PlatformSKU, *pl.QCName, loc)
		if err != nil {
			res.Failed++
			continue
		}
		if err := s.platformPrices.Upsert(itemToPrice(variantID, pl.ID, cityID, item)); err != nil {
			res.Failed++
			continue
		}
		res.Refreshed++
	}
	if err := s.pricing.RecomputeVariantSummary(variantID, cityID); err != nil {
		return res, err
	}
	if cr, err := s.qc.Credits(ctx); err == nil {
		res.CreditsRemaining = cr.Remaining
	}
	return res, nil
}

func (s *LinkingService) RefreshProduct(productID, cityID string) (RefreshResult, error) {
	var res RefreshResult
	vs, err := s.variants.ListByProduct(productID)
	if err != nil {
		return res, err
	}
	for _, v := range vs {
		r, err := s.RefreshVariant(v.ID, cityID)
		if err != nil {
			return res, err
		}
		res.Refreshed += r.Refreshed
		res.Failed += r.Failed
		res.CreditsRemaining = r.CreditsRemaining
	}
	return res, nil
}

func itemToPrice(variantID, platformID, cityID string, item *quickcommerce.ItemDetail) *domain.PlatformPrice {
	p := &domain.PlatformPrice{
		VariantID:  variantID,
		PlatformID: platformID,
		CityID:     cityID,
		PricePaise: item.PricePaise,
		Available:  item.Available,
		Inventory:  item.Inventory,
		Source:     domain.PriceSourceAPI,
	}
	if item.MRPPaise > 0 {
		mrp := item.MRPPaise
		p.MRPPaise = &mrp
	}
	return p
}
