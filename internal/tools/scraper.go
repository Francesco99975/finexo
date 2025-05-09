package tools

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Francesco99975/finexo/internal/database"
	"github.com/Francesco99975/finexo/internal/helpers"
	"github.com/Francesco99975/finexo/internal/models"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
	"github.com/go-rod/stealth"
	"github.com/labstack/gommon/log"
	"golang.org/x/sync/semaphore"
)

func Scrape(seed string, explicit_exchange *string, manager *models.BrowserManager, sem *semaphore.Weighted, mwg *sync.WaitGroup, discoverer *Discoverer) error {

	if sem != nil && mwg != nil {
		defer mwg.Done()

		err := sem.Acquire(context.Background(), 1) // Limit concurrency
		if err != nil {
			return fmt.Errorf("failed to acquire semaphore while scraping seed (%s): %v", seed, err)
		}
		defer sem.Release(1)
	}

	browser := manager.GetBrowser()
	defer manager.ReleaseBrowser()

	var security models.Security
	ticker, exchange_hint, err := tickerExtractor(seed)
	if err != nil {
		return fmt.Errorf("failed to extract ticker and exchange from seed (%s): %v", seed, err)
	}

	security.Ticker = ticker
	var exchange *models.Exchange
	if exchange_hint != "" {
		exchange, err = models.GetExchangeBySuffixorPrefix(database.DB, exchange_hint, exchange_hint)
		if err != nil {
			return fmt.Errorf("failed to get exchange through SUFFIX or PREFIX for seed (%s): %v", seed, err)
		}
		security.Exchange = exchange.Title
	} else {
		if explicit_exchange != nil {
			exchange, err = models.GetExchangeByTitle(database.DB, *explicit_exchange)
			if err != nil {
				return fmt.Errorf("failed to get exchange for seed (%s): %v", seed, err)
			}
			security.Exchange = exchange.Title
		} else {
			ex, err := findExchangeInPage(ticker, BASE_YAHOO_URL+ticker, browser)
			if err != nil {
				return fmt.Errorf("failed to find exchange in page for seed (%s): %v", seed, err)
			}
			exchange, err = models.GetExchangeByTitle(database.DB, ex)
			if err != nil {
				return fmt.Errorf("failed to get exchange for seed (%s): %v", seed, err)
			}
			security.Exchange = ex
		}
	}

	log.Debugf("Scraping %s:%s", security.Ticker, security.Exchange)

	var yahooScrapingUrl string

	if exchange.Suffix.Valid {
		yahooScrapingUrl = BASE_YAHOO_URL + fmt.Sprintf("%s.%s", security.Ticker, exchange.Suffix.String)
	} else {
		yahooScrapingUrl = BASE_YAHOO_URL + ticker
	}

	var marketbeatScrapingUrl string
	if exchange.Prefix.Valid {
		marketbeatScrapingUrl = BASE_MARKETBEAT_URL + fmt.Sprintf("%s/%s", exchange.Prefix.String, security.Ticker)
	} else {
		marketbeatScrapingUrl = BASE_MARKETBEAT_URL + fmt.Sprintf("%s/%s", exchange.Title, security.Ticker)
	}
	//Adjusting For REITS
	marketbeatScrapingUrl = strings.ReplaceAll(marketbeatScrapingUrl, "-UN", "")

	var dividendHostoryScrapingUrl string
	if exchange.CC != "US" {
		if exchange.CC == "UK" {
			dividendHostoryScrapingUrl = BASE_DIVIDENDHISTORY_URL + fmt.Sprintf("%s/%s", strings.ToLower(exchange.CC), security.Ticker)
		} else {
			if exchange.Suffix.String == "NE" {
				dividendHostoryScrapingUrl = BASE_DIVIDENDHISTORY_URL + fmt.Sprintf("%s/%s", "tsx", security.Ticker)
			} else {
				dividendHostoryScrapingUrl = BASE_DIVIDENDHISTORY_URL + fmt.Sprintf("%s/%s", strings.ToLower(exchange.Title), security.Ticker)
			}

		}
	} else {
		dividendHostoryScrapingUrl = BASE_DIVIDENDHISTORY_URL + security.Ticker
	}
	// Adjusting For REITS
	dividendHostoryScrapingUrl = strings.ReplaceAll(dividendHostoryScrapingUrl, "-UN", ".UN")

	page, err := stealth.Page(browser)
	if err != nil {
		return fmt.Errorf("failed to create initial page while working on seed (%s): %v", seed, err)
	}
	defer page.Close()

	// Set a random User-Agent
	userAgent := getRandomUserAgent()
	page.MustSetUserAgent(&proto.NetworkSetUserAgentOverride{UserAgent: userAgent})

	// Spoof WebGL fingerprinting
	spoofWebGLFingerPrint(page)

	// Spoof Canvas fingerprinting
	spoofCanvasFingerPrint(page)

	var wg sync.WaitGroup
	// Create a context to control the Goroutine
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	// Start random behavior in a separate Goroutine
	wg.Add(1)
	// Start random behavior in a separate Goroutine
	go randomUserBehavior(ctx, page, &wg)
	defer wg.Wait()

	log.Debugf("Scraping MarketBeat at url: ", marketbeatScrapingUrl)
	err = page.Navigate(marketbeatScrapingUrl)
	if err != nil {
		log.Warnf("failed to open page on MarketBeat: %v. For seed %s", err, seed)
	}

	err = page.Timeout(20 * time.Second).WaitLoad()
	if err != nil {
		log.Warnf("failed to wait for page load on MarketBeat: %v. For seed %s", err, seed)
	}
	// disableWebRTC(page)

	log.Debugf("Scraping MarketBeat for %s at exchange %s", security.Ticker, security.Exchange)

	//Scrape MarketBeat
	scrapedMarketBeatDataKeys, uperr := page.Timeout(5 * time.Second).Elements(MB_DATA_KEYS)
	scrapedMarketBeatDataValues, err := page.Timeout(5 * time.Second).Elements(MB_DATA_VALUES)
	if err != nil || uperr != nil || len(scrapedMarketBeatDataKeys) == 0 || len(scrapedMarketBeatDataValues) == 0 {
		log.Warnf("failed to scrape MarketBeat data: %v. For seed %s", err, seed)
	} else {
		scrapedMarketBeatDataKeysArray := helpers.MapSlice(scrapedMarketBeatDataKeys, func(e *rod.Element) string {
			return e.MustText()
		})

		scrapedMarketBeatDataValuesArray := helpers.MapSlice(scrapedMarketBeatDataValues, func(e *rod.Element) string {
			return e.MustText()
		})

		for i := range len(scrapedMarketBeatDataKeysArray) {
			key := strings.ToLower(scrapedMarketBeatDataKeysArray[i])

			log.Debugf("Scraped MarketBeat data: %s = %s", key, scrapedMarketBeatDataValuesArray[i])

			if strings.Contains(key, "sector") {
				security.Sector = models.NullableString{String: scrapedMarketBeatDataValuesArray[i], Valid: true}
			}

			if key == "industry" {
				security.Industry = models.NullableString{String: scrapedMarketBeatDataValuesArray[i], Valid: true}
			}

			if strings.Contains(key, "sub") {
				security.SubIndustry = models.NullableString{String: scrapedMarketBeatDataValuesArray[i], Valid: true}
			}

			if strings.Contains(key, "consensus") {
				security.Consensus = models.NullableString{String: scrapedMarketBeatDataValuesArray[i], Valid: true}
			}

			if strings.Contains(key, "score") {
				scrapedScoreStr := scrapedMarketBeatDataValuesArray[i]
				scrapedScoreStr = helpers.NormalizeFloatStrToIntStr(scrapedScoreStr)

				scrapedScore, err := strconv.Atoi(scrapedScoreStr)
				if err != nil {
					log.Warnf("failed to parse score: %v. For seed %s", err, seed)
				} else {
					security.Score = models.NullableInt{Int64: int64(scrapedScore), Valid: true}
				}

			}

			if strings.Contains(key, "coverage") {
				scrapedCoverageStr := strings.Split(scrapedMarketBeatDataValuesArray[i], " ")[0]

				scrapedCoverage, err := strconv.Atoi(scrapedCoverageStr)
				if err != nil {
					log.Warnf("failed to parse coverage: %v. For seed %s", err, seed)
				} else {
					security.Coverage = models.NullableInt{Int64: int64(scrapedCoverage), Valid: true}
				}
			}

			if strings.Contains(key, "outstanding") {
				scrapedOutstandingStr := scrapedMarketBeatDataValuesArray[i]
				scrapedOutstandingStr = helpers.NormalizeFloatStrToIntStr(scrapedOutstandingStr)

				scrapedOutstanding, err := strconv.ParseInt(scrapedOutstandingStr, 10, 64)
				if err != nil {
					log.Warnf("failed to parse outstanding: %v. For seed %s", err, seed)
				} else {
					security.Outstanding = models.NullableInt{Int64: scrapedOutstanding, Valid: true}
				}

			}

		}
	}

	// Scrape Dividend History
	var dividendScrap models.DividendHistoryScrap
	err = page.Navigate(dividendHostoryScrapingUrl)
	if err != nil {
		log.Warnf("failed to open page on Dividend History: %v. For seed %s", err, seed)
	}

	err = page.Timeout(20 * time.Second).WaitLoad()
	if err != nil {
		log.Warnf("failed to wait for page load on Dividend History: %v. For seed %s", err, seed)
	}
	// disableWebRTC(page)

	log.Debugf("Scraping Dividend History for %s at exchange %s", security.Ticker, security.Exchange)

	paragraphs, err := page.Elements("p")
	if err != nil || len(paragraphs) == 0 {
		log.Warnf("failed to scrape Dividend History: %v. For seed %s", err, seed)
	} else {
		for _, paragraph := range paragraphs {
			pt := paragraph.MustText()
			log.Debugf("Scraped Dividend History data PT: %s", pt)
			paragraphText := strings.ReplaceAll(strings.ToLower(pt), " ", "")
			paragraphText = strings.ReplaceAll(paragraphText, "\n", "")

			if strings.Contains(paragraphText, "payoutratio") && strings.Contains(paragraphText, ":") {
				log.Debugf("Scraped Dividend History data: %s", paragraphText)
				scrapedPrStr := strings.Split(paragraphText, ":")[1]
				scrapedPrStr = helpers.NormalizeFloatStrToIntStr(scrapedPrStr)
				scrapedPr, err := strconv.Atoi(scrapedPrStr)
				if err != nil {
					log.Warnf("failed to parse payout ratio: %v. For seed %s", err, seed)
				} else {
					dividendScrap.Pr = &scrapedPr
				}

			}

			if strings.Contains(paragraphText, "frequency") && strings.Contains(paragraphText, ":") {
				log.Debugf("Scraped Dividend History data: %s", paragraphText)
				freq, err := models.ParseFrequency(strings.Split(paragraphText, ":")[1])
				if err != nil {
					log.Warnf("failed to parse frequency: %v. For seed %s", err, seed)
					freqStr := string(models.FrequencyUnknown)
					dividendScrap.Frequency = &freqStr
				} else {
					freqStr := string(freq)
					dividendScrap.Frequency = &freqStr
				}
			}
		}

		if dividendScrap.Frequency == nil {
			freqStr := string(models.FrequencyUnknown)
			dividendScrap.Frequency = &freqStr
		}
	}

	tableIndex := -1
	payoutDates, err := page.Elements("#dividend_table tr td:nth-child(2)")
	if err != nil {
		log.Warnf("failed to scrape Dividend History: %v. For seed %s", err, seed)
	} else {

		for index, payoutDate := range payoutDates {
			date, err := time.Parse("2006-01-02", payoutDate.MustText())
			if err != nil {
				log.Warnf("failed to parse payout date: %v. For seed %s", err, seed)
				continue
			} else {
				if date.After(time.Now()) {
					tableIndex = index + 1
				} else {
					break
				}
			}
		}
	}

	rows, err := page.Elements("table#dividend_table tr")
	if err != nil {
		log.Warnf("failed to scrape Dividend History: %v. For seed %s", err, seed)
	} else if tableIndex != -1 && len(rows) > tableIndex && len(rows) >= 3 {
		relevantRowStr := rows[tableIndex].MustText()
		log.Debugf("Scraped Dividend History data relevantRowStr: %s", relevantRowStr)
		relevantRowArr := strings.Split(relevantRowStr, "\t")

		scrapedExDividendDate, err := time.Parse("2006-01-02", relevantRowArr[0])
		if err != nil {
			log.Warnf("failed to parse ex-dividend date: %v. For seed %s", err, seed)
		} else {
			dividendScrap.ExDivDate = &scrapedExDividendDate
		}

		scrapedPayoutDate, err := time.Parse("2006-01-02", relevantRowArr[1])
		if err != nil {
			log.Warnf("failed to parse payout date: %v. For seed %s", err, seed)
		} else {
			dividendScrap.PayoutDate = &scrapedPayoutDate
		}

		scrapedLadStr := relevantRowArr[2]
		scrapedLadStr = helpers.NormalizeFloatStrToIntStr(scrapedLadStr)
		if len(scrapedLadStr) >= 3 {
			scrapedLadStr = scrapedLadStr[:3]
		}
		scrapedLad, err := strconv.Atoi(scrapedLadStr)
		if err != nil {
			log.Warnf("failed to parse lad: %v. For seed %s", err, seed)
		} else {
			dividendScrap.Lad = &scrapedLad
		}

	}

	err = page.Navigate(yahooScrapingUrl)
	if err != nil {
		return fmt.Errorf("failed to open page on Yahoo: %v. For seed %s", err, seed)
	}

	err = page.Timeout(20 * time.Second).WaitLoad()
	if err != nil {
		return fmt.Errorf("failed to wait for page load on Yahoo: %v. For seed %s", err, seed)
	}
	// disableWebRTC(page)

	if discoverer != nil {

		scrapedDiscoveredSeeds, err := page.Timeout(5 * time.Second).Elements(YH_DISCOVER_SEEDS_SELECTOR)
		if err != nil {
			log.Warnf("failed to scrape discovered seeds: %v. For seed %s", err, seed)
		}

		log.Debugf("Scraping Yahoo reccomanded seeds for %s at exchange %s -- Found %d seeds", security.Ticker, security.Exchange, len(scrapedDiscoveredSeeds))

		for _, discoveredSeed := range scrapedDiscoveredSeeds {
			seed, err := discoveredSeed.Attribute("title")
			if err != nil {
				log.Warnf("failed to scrape discovered seed url: %v. For seed %s", err, seed)
			}

			err = discoverer.Collect(*seed)
			if err != nil {
				log.Warnf("failed to collect discovered seed: %v. For seed %s", err, seed)
			}
		}
		log.Debugf("Collected Yahoo reccomanded seeds for %s at exchange %s", security.Ticker, security.Exchange)
	}

	scrapedCurrencyElem, err := page.Timeout(5 * time.Second).Element(YH_CURRENCY_SELECTOR)
	if err != nil {
		return fmt.Errorf("currency not found in page - target: %s:%s", security.Ticker, security.Exchange)
	}

	scrapedCurrency := scrapedCurrencyElem.MustText()

	log.Debugf("Scraped currency: %s", scrapedCurrency)

	scrapedCurrency = strings.TrimSpace(scrapedCurrency)

	if isAnEmptyString(scrapedCurrency) {
		return fmt.Errorf("empty currency: %s - target: %s:%s", scrapedCurrency, security.Ticker, security.Exchange)
	}

	security.Currency = scrapedCurrency

	scrapedFullNameElem, err := page.Timeout(5 * time.Second).Element(YH_FULLNAME_SELECTOR)
	if err != nil {
		return fmt.Errorf("full name not found in page - target: %s:%s", security.Ticker, security.Exchange)
	}

	scrapedFullName := scrapedFullNameElem.MustText()

	if isAnEmptyString(scrapedFullName) {
		return fmt.Errorf("empty full name: %s - target: %s:%s", scrapedFullName, security.Ticker, security.Exchange)
	}

	security.FullName = scrapedFullName
	log.Debug("Scraped full name")

	scrapedTypologyREITHint, err := page.Timeout(5 * time.Second).Element(YH_REIT_HINT_SELECTOR)
	if err != nil {
		return fmt.Errorf("typology hint not found in page - target: %s:%s", security.Ticker, security.Exchange)
	}

	scrapedTypologyREITHintStr, err := scrapedTypologyREITHint.Text()
	if err != nil {
		return fmt.Errorf("failed to get typology hint text: %v. For seed %s", err, seed)
	}

	scrapdTypologyETFHint, err := page.Timeout(5 * time.Second).Element(YH_ETF_HINT_SELECTOR)
	if err != nil {
		return fmt.Errorf("typology ETF hint not found in page - target: %s:%s", security.Ticker, security.Exchange)
	}

	scrapedTypologyETFHintStr, err := scrapdTypologyETFHint.Text()
	if err != nil {
		return fmt.Errorf("failed to get typology ETF hint text: %v. For seed %s", err, seed)
	}

	scrapedTypology := "STOCK"
	if strings.Contains(strings.ToLower(scrapedTypologyREITHintStr), "reit") {
		scrapedTypology = "REIT"
	}

	if strings.Contains(strings.ToLower(scrapedTypologyETFHintStr), "fund family") {
		scrapedTypology = "ETF"
	}

	security.Typology = scrapedTypology
	log.Debugf("Scraped typology: %s", scrapedTypology)

	priceStrElem, err := page.Timeout(5 * time.Second).Element(YH_PRICE_SELECTOR)
	if err != nil {
		return fmt.Errorf("price not found in page - target: %s:%s", security.Ticker, security.Exchange)
	}

	priceStr := priceStrElem.MustText()
	log.Debugf("Scraped price: %s", priceStr)
	priceStr = helpers.NormalizeFloatStrToIntStr(priceStr)

	if isAnEmptyString(priceStr) {
		return fmt.Errorf("empty price: %s - target: %s:%s", priceStr, security.Ticker, security.Exchange)
	}

	scrapedPrice, err := strconv.Atoi(priceStr)
	if err != nil {
		return fmt.Errorf("invalid price: %s - target: %s:%s", priceStr, security.Ticker, security.Exchange)
	}

	if scrapedPrice <= 0 {
		return fmt.Errorf("invalid negative price: %d - target: %s:%s", scrapedPrice, security.Ticker, security.Exchange)
	}

	security.Price = scrapedPrice

	priceChangeStrElem, err := page.Timeout(5 * time.Second).Element(YH_PCHANGE_SELECTOR)
	if err != nil {
		return fmt.Errorf("price change not found in page - target: %s:%s", security.Ticker, security.Exchange)
	}
	priceChangeStr := priceChangeStrElem.MustText()
	log.Debugf("Scraped price change: %s", priceChangeStr)
	priceChangeStr = helpers.NormalizeFloatStrToIntStr(priceChangeStr)

	if isAnEmptyString(priceChangeStr) {
		return fmt.Errorf("empty price change: %s - target: %s:%s", priceChangeStr, security.Ticker, security.Exchange)
	}

	scrapedPriceChange, err := strconv.Atoi(priceChangeStr)
	if err != nil {
		return fmt.Errorf("invalid price change: %s - target: %s:%s", priceChangeStr, security.Ticker, security.Exchange)
	}

	security.PC = scrapedPriceChange
	log.Debug("Scraped price change")

	priceChangePercentageStrElem, err := page.Timeout(5 * time.Second).Element(YH_PRICE_PERCENTAGE_CHANGE_SELECTOR)
	if err != nil {
		return fmt.Errorf("price change percentage not found in page - target: %s:%s", security.Ticker, security.Exchange)
	}

	priceChangePercentageStr := priceChangePercentageStrElem.MustText()
	log.Debugf("Scraped price change percentage: %s", priceChangePercentageStr)
	priceChangePercentageStr = helpers.NormalizeFloatStrToIntStr(priceChangePercentageStr)

	if isAnEmptyString(priceChangePercentageStr) {
		return fmt.Errorf("empty price change percentage: %s - target: %s:%s", priceChangePercentageStr, security.Ticker, security.Exchange)
	}

	scrapedPriceChangePercentage, err := strconv.Atoi(priceChangePercentageStr)
	if err != nil {
		return fmt.Errorf("invalid price change percentage: %s - target: %s:%s", priceChangePercentageStr, security.Ticker, security.Exchange)
	}

	security.PCP = scrapedPriceChangePercentage

	yearlyRangeStrElem, err := page.Timeout(5 * time.Second).Element(YH_YEARLY_RANGE_SELECTOR)
	if err != nil {
		return fmt.Errorf("yearly range not found in page - target: %s:%s", security.Ticker, security.Exchange)
	}

	yearlyRangeStr := yearlyRangeStrElem.MustText()
	log.Debugf("Scraped yearly range: %s", yearlyRangeStr)

	yearlyRangeStr = strings.ReplaceAll(yearlyRangeStr, " ", "")
	yearlyRangeArr := strings.Split(yearlyRangeStr, "-")

	if len(yearlyRangeArr) != 2 {
		return fmt.Errorf("invalid yearly range: %s - target: %s:%s", yearlyRangeStr, security.Ticker, security.Exchange)
	}

	yrlStr := yearlyRangeArr[0]
	yrlStr = helpers.NormalizeFloatStrToIntStr(yrlStr)
	if yrlStr == "" {
		return fmt.Errorf("empty yearly range low: %s - target: %s:%s", yrlStr, security.Ticker, security.Exchange)
	}

	scrapedYrl, err := strconv.Atoi(yrlStr)
	if err != nil {
		return fmt.Errorf("invalid yearly range low: %s - target: %s:%s", yrlStr, security.Ticker, security.Exchange)
	}

	if scrapedYrl <= 0 {
		return fmt.Errorf("invalid negative yearly range low: %d - target: %s:%s", scrapedYrl, security.Ticker, security.Exchange)
	}

	security.YearLow = scrapedYrl
	log.Debug("Scraped yearly range low")

	yrhStr := yearlyRangeArr[1]
	yrhStr = helpers.NormalizeFloatStrToIntStr(yrhStr)
	if yrhStr == "" {
		return fmt.Errorf("empty yearly range high: %s - target: %s:%s", yrhStr, security.Ticker, security.Exchange)
	}

	scrapedYrh, err := strconv.Atoi(yrhStr)
	if err != nil {
		return fmt.Errorf("invalid yearly range high: %s - target: %s:%s", yrhStr, security.Ticker, security.Exchange)
	}

	if scrapedYrh <= 0 {
		return fmt.Errorf("invalid negative yearly range high: %d - target: %s:%s", scrapedYrh, security.Ticker, security.Exchange)
	}

	if scrapedYrh < scrapedYrl {
		return fmt.Errorf("invalid yearly range high < low: %d < %d - target: %s:%s", scrapedYrh, scrapedYrl, security.Ticker, security.Exchange)
	}

	security.YearHigh = scrapedYrh
	log.Debug("Scraped yearly range high")

	daylyRangeStrElem, err := page.Timeout(5 * time.Second).Element(YH_DAILY_RANGE_SELECTOR)
	if err != nil {
		return fmt.Errorf("daily range not found in page - target: %s:%s", security.Ticker, security.Exchange)
	}

	daylyRangeStr := daylyRangeStrElem.MustText()
	log.Debugf("Scraped daily range: %s", daylyRangeStr)

	daylyRangeStr = strings.ReplaceAll(daylyRangeStr, " ", "")
	daylyRangeArr := strings.Split(daylyRangeStr, "-")
	if len(daylyRangeArr) != 2 {
		return fmt.Errorf("invalid daily range: %s - target: %s:%s", daylyRangeStr, security.Ticker, security.Exchange)
	}

	drlStr := daylyRangeArr[0]
	drlStr = helpers.NormalizeFloatStrToIntStr(drlStr)
	if drlStr == "" {
		return fmt.Errorf("empty daily range low: %s - target: %s:%s", drlStr, security.Ticker, security.Exchange)
	}

	scrapedDrl, err := strconv.Atoi(drlStr)
	if err != nil {
		return fmt.Errorf("invalid daily range low: %s - target: %s:%s", drlStr, security.Ticker, security.Exchange)
	}

	if scrapedDrl <= 0 {
		return fmt.Errorf("invalid negative daily range low: %d - target: %s:%s", scrapedDrl, security.Ticker, security.Exchange)
	}

	security.DayLow = scrapedDrl
	log.Debug("Scraped daily range low")

	drhStr := daylyRangeArr[1]
	drhStr = helpers.NormalizeFloatStrToIntStr(drhStr)
	if drhStr == "" {
		return fmt.Errorf("empty daily range high: %s - target: %s:%s", drhStr, security.Ticker, security.Exchange)
	}

	scrapedDrh, err := strconv.Atoi(drhStr)
	if err != nil {
		return fmt.Errorf("invalid daily range high: %s - target: %s:%s", drhStr, security.Ticker, security.Exchange)
	}

	if scrapedDrh <= 0 {
		return fmt.Errorf("invalid negative daily range high: %d - target: %s:%s", scrapedDrh, security.Ticker, security.Exchange)
	}

	if scrapedDrh < scrapedDrl {
		return fmt.Errorf("invalid daily range high < low: %d < %d - target: %s:%s", scrapedDrh, scrapedDrl, security.Ticker, security.Exchange)
	}

	security.DayHigh = scrapedDrh
	log.Debug("Scraped daily range high")

	marketCapStrElem, err := page.Timeout(5 * time.Second).Element(YH_MARKET_CAP_SELECTOR)
	if err != nil {
		log.Warnf("market cap not found in page - target: %s:%s", security.Ticker, security.Exchange)
		security.MarketCap = models.NullableInt{
			Valid: false,
		}
	} else {
		marketCapStr := marketCapStrElem.MustText()
		log.Debugf("Scraped market cap: %s", marketCapStr)

		if isAnEmptyString(marketCapStr) {
			log.Warnf("empty market cap: %s - target: %s:%s", marketCapStr, security.Ticker, security.Exchange)
			security.MarketCap = models.NullableInt{
				Valid: false,
			}
		} else {
			scrapedMarketCap, err := helpers.ParseNumberString(marketCapStr)
			if err != nil || scrapedMarketCap <= 0 {
				log.Warnf("invalid market cap: %s - target: %s:%s", marketCapStr, security.Ticker, security.Exchange)
				security.MarketCap = models.NullableInt{
					Valid: false,
				}
			} else {
				security.MarketCap = models.NullableInt{
					Int64: scrapedMarketCap,
					Valid: true,
				}
			}
		}
	}

	volumeStrElem, err := page.Timeout(5 * time.Second).Element(YH_VOLUME_SELECTOR)
	if err != nil {
		log.Warnf("volume not found in page - target: %s:%s", security.Ticker, security.Exchange)
		security.Volume = models.NullableInt{
			Valid: false,
		}
	} else {
		volumeStr := volumeStrElem.MustText()
		log.Debugf("Scraped volume: %s", volumeStr)
		volumeStr = strings.ReplaceAll(volumeStr, ",", "")

		if isAnEmptyString(volumeStr) {
			log.Warnf("empty volume: %s - target: %s:%s", volumeStr, security.Ticker, security.Exchange)
			security.Volume = models.NullableInt{
				Valid: false,
			}
		} else {
			scrapedVolume, err := strconv.ParseInt(volumeStr, 10, 64)
			if err != nil || scrapedVolume <= 0 {
				log.Warnf("invalid volume: %s - target: %s:%s", volumeStr, security.Ticker, security.Exchange)
				security.Volume = models.NullableInt{
					Valid: false,
				}
			} else {
				security.Volume = models.NullableInt{
					Int64: scrapedVolume,
					Valid: true,
				}
			}
		}
	}

	log.Debug("Scraped volume")

	avgVolumeStrElem, err := page.Timeout(5 * time.Second).Element(YH_AVG_VOLUME_SELECTOR)
	if err != nil {
		log.Warnf("average volume not found in page - target: %s:%s", security.Ticker, security.Exchange)
		security.AvgVolume = models.NullableInt{
			Valid: false,
		}
	} else {
		avgVolumeStr := avgVolumeStrElem.MustText()
		log.Debugf("Scraped average volume: %s", avgVolumeStr)
		avgVolumeStr = strings.ReplaceAll(avgVolumeStr, ",", "")

		if isAnEmptyString(avgVolumeStr) {
			log.Warnf("empty average volume: %s - target: %s:%s", avgVolumeStr, security.Ticker, security.Exchange)
			security.AvgVolume = models.NullableInt{
				Valid: false,
			}
		} else {
			scrapedAvgVolume, err := strconv.ParseInt(avgVolumeStr, 10, 64)
			if err != nil || scrapedAvgVolume <= 0 {
				log.Warnf("invalid average volume: %s - target: %s:%s", avgVolumeStr, security.Ticker, security.Exchange)
				security.AvgVolume = models.NullableInt{
					Valid: false,
				}
			} else {
				security.AvgVolume = models.NullableInt{
					Int64: scrapedAvgVolume,
					Valid: true,
				}
			}
		}

	}

	log.Debug("Scraped average volume")

	betaStrElem, err := page.Timeout(5 * time.Second).Element(YH_BETA_SELECTOR)
	if err != nil {
		log.Warnf("beta not found in page - target: %s:%s", security.Ticker, security.Exchange)
		security.Beta = models.NullableInt{
			Valid: false,
		}
	} else {
		betaStr := betaStrElem.MustText()
		log.Debugf("Scraped beta: %s", betaStr)
		betaStr = helpers.NormalizeFloatStrToIntStr(betaStr)
		if isAnEmptyString(betaStr) {
			log.Warnf("empty beta: %s - target: %s:%s", betaStr, security.Ticker, security.Exchange)
			security.Beta = models.NullableInt{
				Valid: false,
			}
		} else {
			scrapedBeta, err := strconv.Atoi(betaStr)
			if err != nil {
				log.Warnf("invalid beta: %s - target: %s:%s", betaStr, security.Ticker, security.Exchange)
				security.Beta = models.NullableInt{
					Valid: false,
				}
			} else {
				security.Beta = models.NullableInt{
					Int64: int64(scrapedBeta),
					Valid: true,
				}
			}
		}
	}

	log.Debug("Scraped beta")

	pcloseStrElem, err := page.Timeout(5 * time.Second).Element(YH_PCLOSE_SELECTOR)
	if err != nil {
		return fmt.Errorf("previous close not found in page - target: %s:%s", security.Ticker, security.Exchange)
	}

	pcloseStr := pcloseStrElem.MustText()
	log.Debugf("Scraped previous close: %s", pcloseStr)
	pcloseStr = helpers.NormalizeFloatStrToIntStr(pcloseStr)
	if isAnEmptyString(pcloseStr) {
		return fmt.Errorf("empty previous close: %s - target: %s:%s", pcloseStr, security.Ticker, security.Exchange)
	}

	scrapedPclose, err := strconv.Atoi(pcloseStr)
	if err != nil {
		return fmt.Errorf("invalid previous close: %s - target: %s:%s", pcloseStr, security.Ticker, security.Exchange)
	}

	if scrapedPclose <= 0 {
		return fmt.Errorf("invalid negative previous close: %d - target: %s:%s", scrapedPclose, security.Ticker, security.Exchange)
	}

	security.PClose = scrapedPclose
	log.Debug("Scraped previous close")

	targetStrElem, err := page.Timeout(5 * time.Second).Element(YH_TARGET_SELECTOR)
	if err != nil {
		log.Warnf("target not found in page - target: %s:%s", security.Ticker, security.Exchange)
	} else {
		targetStr := targetStrElem.MustText()
		log.Debugf("Scraped target: %s", targetStr)
		targetStr = helpers.NormalizeFloatStrToIntStr(targetStr)
		if isAnEmptyString(targetStr) {
			log.Warnf("empty target: %s - target: %s:%s", targetStr, security.Ticker, security.Exchange)
		} else {
			scrapedTarget, err := strconv.Atoi(targetStr)
			if err != nil || scrapedTarget <= 0 {
				log.Warnf("invalid target: %s - target: %s:%s", targetStr, security.Ticker, security.Exchange)
			} else {
				security.Target = models.NullableInt{
					Int64: int64(scrapedTarget),
					Valid: true,
				}
			}
		}
	}

	copenStrElem, err := page.Timeout(5 * time.Second).Element(YH_COPEN_SELECTOR)
	if err != nil {
		return fmt.Errorf("open not found in page - target: %s:%s", security.Ticker, security.Exchange)
	}
	copenStr := copenStrElem.MustText()
	log.Debugf("Scraped open: %s", copenStr)
	copenStr = helpers.NormalizeFloatStrToIntStr(copenStr)
	if isAnEmptyString(copenStr) {
		return fmt.Errorf("empty open: %s - target: %s:%s", copenStr, security.Ticker, security.Exchange)
	}

	scrapedCopen, err := strconv.Atoi(copenStr)
	if err != nil {
		return fmt.Errorf("invalid open: %s - target: %s:%s", copenStr, security.Ticker, security.Exchange)
	}

	if scrapedCopen <= 0 {
		return fmt.Errorf("invalid negative open: %d - target: %s:%s", scrapedCopen, security.Ticker, security.Exchange)
	}

	security.COpen = scrapedCopen
	log.Debug("Scraped open")

	bidPayloadStrElem, err := page.Timeout(5 * time.Second).Element(YH_BID_SELECTOR)
	if err != nil {
		return fmt.Errorf("bid not found in page - target: %s:%s", security.Ticker, security.Exchange)
	}

	bidPayloadStr := bidPayloadStrElem.MustText()
	log.Debugf("Scraped bid: %s", bidPayloadStr)
	bidPayloadStr = strings.ReplaceAll(bidPayloadStr, " ", "")
	bidPayloadArr := strings.Split(bidPayloadStr, "x")

	if isAnEmptyString(bidPayloadStr) || len(bidPayloadArr) != 2 {
		security.Bid = security.Price
	} else {
		bidStr := bidPayloadArr[0]
		bidStr = helpers.NormalizeFloatStrToIntStr(bidStr)
		if bidStr == "" {
			return fmt.Errorf("empty bid: %s - target: %s:%s", bidStr, security.Ticker, security.Exchange)
		}

		scrapedBid, err := strconv.Atoi(bidStr)
		if err != nil {
			return fmt.Errorf("invalid bid: %s - target: %s:%s", bidStr, security.Ticker, security.Exchange)
		}

		if scrapedBid <= 0 {
			return fmt.Errorf("invalid negative bid: %d - target: %s:%s", scrapedBid, security.Ticker, security.Exchange)
		}

		security.Bid = scrapedBid
		log.Debug("Scraped bid")

		bidSizeStr := bidPayloadArr[1]
		if isAnEmptyString(bidSizeStr) {
			log.Warnf("empty bid size: %s - target: %s:%s", bidSizeStr, security.Ticker, security.Exchange)
			security.BidSize = models.NullableInt{
				Valid: false,
			}
		} else {
			scrapedBidSize, err := strconv.Atoi(bidSizeStr)
			if err != nil || scrapedBidSize < 0 {
				log.Warnf("invalid bid size: %s - target: %s:%s", bidSizeStr, security.Ticker, security.Exchange)
				security.BidSize = models.NullableInt{
					Valid: false,
				}
			} else {
				security.BidSize = models.NullableInt{
					Int64: int64(scrapedBidSize),
					Valid: true,
				}
			}

		}
		log.Debug("Scraped bid size")
	}

	askPayloadStrElem, err := page.Timeout(5 * time.Second).Element(YH_ASK_SELECTOR)
	if err != nil {
		return fmt.Errorf("ask not found in page - target: %s:%s", security.Ticker, security.Exchange)
	}

	askPayloadStr := askPayloadStrElem.MustText()
	log.Debugf("Scraped ask: %s", askPayloadStr)
	askPayloadStr = strings.ReplaceAll(askPayloadStr, " ", "")
	askPayloadArr := strings.Split(askPayloadStr, "x")

	if isAnEmptyString(askPayloadStr) || len(askPayloadArr) != 2 {
		security.Ask = security.Price
	} else {

		askStr := askPayloadArr[0]
		askStr = helpers.NormalizeFloatStrToIntStr(askStr)
		if isAnEmptyString(askStr) {
			return fmt.Errorf("empty ask: %s - target: %s:%s", askStr, security.Ticker, security.Exchange)
		}

		scrapedAsk, err := strconv.Atoi(askStr)
		if err != nil {
			return fmt.Errorf("invalid ask: %s - target: %s:%s", askStr, security.Ticker, security.Exchange)
		}

		if scrapedAsk <= 0 {
			return fmt.Errorf("invalid negative ask: %d - target: %s:%s", scrapedAsk, security.Ticker, security.Exchange)
		}

		security.Ask = scrapedAsk
		log.Debug("Scraped ask")

		askSizeStr := askPayloadArr[1]
		if isAnEmptyString(askSizeStr) {
			log.Warnf("empty ask size: %s - target: %s:%s", askSizeStr, security.Ticker, security.Exchange)
			security.AskSize = models.NullableInt{
				Valid: false,
			}
		} else {
			scrapedAskSize, err := strconv.Atoi(askSizeStr)
			if err != nil || scrapedAskSize <= 0 {
				log.Warnf("invalid ask size: %s - target: %s:%s", askSizeStr, security.Ticker, security.Exchange)
				security.AskSize = models.NullableInt{
					Valid: false,
				}
			}

			security.AskSize = models.NullableInt{
				Int64: int64(scrapedAskSize),
				Valid: true,
			}
		}

		log.Debug("Scraped ask size")
	}

	stockDataElements, err := page.Timeout(5 * time.Second).Elements(YH_STOCK_DATA_SELECTOR)
	if err != nil {
		return fmt.Errorf("trailing PE not found in page - target: %s:%s", security.Ticker, security.Exchange)
	}

	if len(stockDataElements) == 0 {
		log.Warnf("empty trailing PE: %s - target: %s:%s", stockDataElements, security.Ticker, security.Exchange)
		security.PE = models.NullableInt{
			Valid: false,
		}
	}

	if len(stockDataElements) == 1 || len(stockDataElements) == 2 {
		peStr := stockDataElements[0].MustText()
		log.Debugf("Scraped trailing PE: %s", peStr)
		peStr = helpers.NormalizeFloatStrToIntStr(peStr)
		if peStr == "" {
			return fmt.Errorf("empty trailing PE: %s - target: %s:%s", peStr, security.Ticker, security.Exchange)
		}

		scrapedPe, err := strconv.Atoi(peStr)
		if err != nil || scrapedPe <= 0 {
			log.Warnf("invalid trailing PE: %s - target: %s:%s", peStr, security.Ticker, security.Exchange)
			security.PE = models.NullableInt{
				Valid: false,
			}
		} else {
			security.PE = models.NullableInt{
				Int64: int64(scrapedPe),
				Valid: true,
			}
		}
	}
	log.Debug("Scraped trailing PE")

	if len(stockDataElements) == 2 {
		epsStr := stockDataElements[1].MustText()
		log.Debugf("Scraped EPS: %s", epsStr)
		epsStr = helpers.NormalizeFloatStrToIntStr(epsStr)
		if epsStr == "" {
			return fmt.Errorf("empty EPS: %s - target: %s:%s", epsStr, security.Ticker, security.Exchange)
		}

		scrapedEps, err := strconv.Atoi(epsStr)
		if err != nil {
			log.Warnf("invalid EPS: %s - target: %s:%s", epsStr, security.Ticker, security.Exchange)
			security.EPS = models.NullableInt{
				Valid: false,
			}
		} else {
			security.EPS = models.NullableInt{
				Int64: int64(scrapedEps),
				Valid: true,
			}
		}
	}

	log.Debug("Scraped EPS")

	security.STM = models.NullableString{
		String: string(models.TimingTTM),
		Valid:  true,
	}

	security.Dividend = scrapeDividend(ticker, security.Exchange, security.Typology, page)
	log.Debug("Scraped dividend")

	if security.Dividend != nil {
		if dividendScrap.Lad != nil {
			security.Dividend.LastAnnounced = models.NullableInt{
				Int64: int64(*dividendScrap.Lad),
				Valid: true,
			}
		}

		if dividendScrap.Pr != nil {
			security.Dividend.PayoutRatio = models.NullableInt{
				Int64: int64(*dividendScrap.Pr),
				Valid: true,
			}
		}

		// if scrapedSeekingAlphaData.Lgr != nil {
		// 	security.Dividend.GrowthRate = models.NullableInt{
		// 		Int64: int64(*scrapedSeekingAlphaData.Lgr),
		// 		Valid: true,
		// 	}
		// }

		// if scrapedSeekingAlphaData.Yog != nil {
		// 	security.Dividend.YearsGrowth = models.NullableInt{
		// 		Int64: int64(*scrapedSeekingAlphaData.Yog),
		// 		Valid: true,
		// 	}
		// }

		if dividendScrap.Frequency != nil {
			security.Dividend.Frequency = models.NullableString{
				String: string(models.Frequency(*dividendScrap.Frequency)),
				Valid:  true,
			}

		}

		if security.Dividend.Frequency.Valid && security.Dividend.LastAnnounced.Valid {
			if security.Dividend.Frequency.String != string(models.FrequencyUnknown) {
				security.Dividend.AnnualPayout = models.NullableInt{
					Int64: int64(int(security.Dividend.LastAnnounced.Int64) * operandByFrequency(dividendScrap.Frequency)),
					Valid: true,
				}
			} else {
				security.Dividend.AnnualPayout = models.NullableInt{
					Int64: int64(math.Floor(float64(security.Price) * (float64(security.Dividend.Yield) / 100) / 100)),
					Valid: true,
				}
			}
		}

		if dividendScrap.ExDivDate != nil {
			security.Dividend.ExDivDate = models.NullableTime{
				Time:  time.Time(*dividendScrap.ExDivDate),
				Valid: true,
			}
		}

		if dividendScrap.PayoutDate != nil {
			security.Dividend.PayoutDate = models.NullableTime{
				Time:  time.Time(*dividendScrap.PayoutDate),
				Valid: true,
			}
		}

	}

	switch security.Typology {
	case "STOCK":

		// Check if security already exists in DB
		exists := models.SecurityExists(database.DB, security.Ticker, security.Exchange)
		if !exists {
			start := time.Now()
			err = models.CreateStock(database.DB, &security)
			if err != nil {
				return fmt.Errorf("error creating stock: %v", err)
			}
			log.Infof("Created Stock based on Scraped data: %v", security.CreatePrettyPrintString())
			helpers.RecordDBQueryLatency("create_stock", start)
			helpers.RecordBusinessEvent("stock_created")
			helpers.RecordBusinessEvent("security_created")
		} else {
			start := time.Now()
			err = models.UpdateStock(database.DB, &security)
			if err != nil {
				return fmt.Errorf("error updating stock: %v", err)
			}
			log.Infof("Updated Stock based on Scraped data: %v", security.CreatePrettyPrintString())
			helpers.RecordDBQueryLatency("update_stock", start)
			helpers.RecordBusinessEvent("stock_updated")
			helpers.RecordBusinessEvent("security_updated")
		}
	case "ETF":
		var etf models.ETF

		etf.Security = security

		aumStrElem, err := page.Timeout(5 * time.Second).Element(YH_AUM_SELECTOR)
		if err != nil {
			log.Warnf("AUM not found in page - target: %s:%s", security.Ticker, security.Exchange)
			etf.AUM = models.NullableInt{
				Valid: false,
			}
		} else {
			aumStr := aumStrElem.MustText()
			log.Debugf("Scraped AUM: %s", aumStr)
			scrapedAum, err := helpers.ParseNumberString(aumStr)
			if err != nil || scrapedAum <= 0 {
				log.Warnf("invalid AUM: %s - target: %s:%s", aumStr, security.Ticker, security.Exchange)
				etf.AUM = models.NullableInt{
					Valid: false,
				}
			} else {
				etf.AUM = models.NullableInt{
					Int64: scrapedAum,
					Valid: true,
				}
			}
		}

		log.Debug("Scraped AUM")

		erStrElem, err := page.Timeout(5 * time.Second).Element(YH_ER_SELECTOR)
		if err != nil {
			log.Warnf("expense ratio not found in page - target: %s:%s", security.Ticker, security.Exchange)
			etf.ExpenseRatio = models.NullableInt{
				Valid: false,
			}
		} else {
			erStr := erStrElem.MustText()
			log.Debugf("Scraped expense ratio: %s", erStr)
			erStr = helpers.NormalizeFloatStrToIntStr(erStr)
			if erStr == "" {
				log.Warnf("empty expense ratio: %s - target: %s:%s", erStr, security.Ticker, security.Exchange)
				etf.ExpenseRatio = models.NullableInt{
					Valid: false,
				}
			} else {
				scrapedEr, err := strconv.Atoi(erStr)
				if err != nil || scrapedEr <= 0 {
					log.Warnf("invalid expense ratio: %s - target: %s:%s", erStr, security.Ticker, security.Exchange)
					etf.ExpenseRatio = models.NullableInt{
						Valid: false,
					}
				} else {
					etf.ExpenseRatio = models.NullableInt{
						Int64: int64(scrapedEr),
						Valid: true,
					}
				}
			}
		}

		log.Debug("Scraped expense ratio")

		navStrElem, err := page.Timeout(5 * time.Second).Element(YH_NAV_SELECTOR)
		if err != nil {
			log.Warnf("NAV not found in page - target: %s:%s", security.Ticker, security.Exchange)
			etf.NAV = models.NullableInt{
				Valid: false,
			}
		} else {
			navStr := navStrElem.MustText()
			log.Debugf("Scraped NAV: %s", navStr)
			navStr = helpers.NormalizeFloatStrToIntStr(navStr)
			if isAnEmptyString(navStr) {
				log.Warnf("empty NAV: %s - target: %s:%s", navStr, security.Ticker, security.Exchange)
				etf.NAV = models.NullableInt{
					Valid: false,
				}
			} else {
				scrapedNav, err := strconv.Atoi(navStr)
				if err != nil || scrapedNav <= 0 {
					log.Warnf("invalid NAV: %s - target: %s:%s", navStr, security.Ticker, security.Exchange)
					etf.NAV = models.NullableInt{
						Valid: false,
					}
				} else {
					etf.NAV = models.NullableInt{
						Int64: int64(scrapedNav),
						Valid: true,
					}
				}
			}
		}

		log.Debug("Scraped NAV")

		EtfDataElems, err := page.Timeout(5 * time.Second).Elements(YH_ETF_DATA_SELECTOR)
		if err != nil {
			log.Warnf("inception date not found in page - target: %s:%s", security.Ticker, security.Exchange)
			etf.InceptionDate = models.NullableTime{
				Valid: false,
			}
		} else {
			family := EtfDataElems[0].MustText()
			log.Debugf("Scraped family: %s", family)
			etf.Family = family

			inceptionDateStr := EtfDataElems[3].MustText()
			log.Debugf("Scraped inception date: %s", inceptionDateStr)
			if isAnEmptyString(inceptionDateStr) {
				log.Warnf("empty inception date: %s - target: %s:%s", inceptionDateStr, security.Ticker, security.Exchange)
				etf.InceptionDate = models.NullableTime{
					Valid: false,
				}
			} else {
				scrapedInceptionDate, err := time.Parse("2006-01-02", inceptionDateStr)
				if err != nil {
					log.Warnf("invalid inception date: %s - target: %s:%s", inceptionDateStr, security.Ticker, security.Exchange)
					etf.InceptionDate = models.NullableTime{
						Valid: false,
					}
				} else {
					etf.InceptionDate = models.NullableTime{
						Time:  scrapedInceptionDate,
						Valid: true,
					}
				}
			}
		}

		log.Debug("Scraped inception date")

		relationsElementsTickers, err := page.Timeout(5 * time.Second).Elements(YH_HOLDINGS_TICKERS_SELECTOR)
		if err != nil {
			log.Warnf("top holdings not found in page - target: %s:%s", security.Ticker, security.Exchange)
		}

		relationsElementsAllocations, err := page.Timeout(5 * time.Second).Elements(YH_HOLDING_ALLOCATIONS_SELECTOR)
		if err != nil {
			log.Warnf("top holdings not found in page - target: %s:%s", security.Ticker, security.Exchange)
		}

		relationsElementsTickersArr := helpers.MapSlice(relationsElementsTickers, func(elem *rod.Element) string {
			return elem.MustText()
		})
		relationsElementsAllocationsArr := helpers.MapSlice(relationsElementsAllocations, func(elem *rod.Element) string {
			return elem.MustText()
		})

		log.Debugf("Scraped top holdings: %v", relationsElementsTickersArr)
		log.Debugf("Scraped top holdings allocations: %v", relationsElementsAllocationsArr)

		var gapSums []float64
		for i := range len(relationsElementsTickersArr) {
			seed := relationsElementsTickersArr[i]
			log.Debugf("Scraped top holding: %s", seed)
			seed = strings.TrimSpace(seed)
			if isAnEmptyString(seed) {
				log.Warnf("empty top holding: %s - target: %s:%s", seed, security.Ticker, security.Exchange)
				continue
			}

			relatedTicker, relatedExchange, err := tickerExtractor(seed)
			if err != nil {
				log.Warnf("invalid top holding: %s - target: %s:%s", relatedTicker, security.Ticker, security.Exchange)
				continue
			}
			if isAnEmptyString(relatedTicker) {
				log.Warnf("empty top holding: %s - target: %s:%s", relatedTicker, security.Ticker, security.Exchange)
				continue
			}
			allocationStr := relationsElementsAllocationsArr[i]
			log.Debugf("Scraped top holding allocation: %s", allocationStr)
			allocationStr = helpers.NormalizeFloatStrToIntStr(allocationStr)
			if isAnEmptyString(allocationStr) {
				log.Warnf("empty allocation: %s - target: %s:%s", allocationStr, security.Ticker, security.Exchange)
				continue
			}

			scrapedAllocation, err := strconv.Atoi(allocationStr)
			if err != nil || scrapedAllocation <= 0 {
				log.Warnf("invalid allocation: %s - target: %s:%s", allocationStr, security.Ticker, security.Exchange)
				continue
			}

			//Steps to find related exchange
			var relatedExchangeInfo *models.Exchange
			if relatedExchange == "" {
				relatedExchange, err = findExchangeInPage(ticker, BASE_YAHOO_URL+relatedTicker, browser)
				if err != nil {
					log.Warnf("invalid exchange or could not find: %s - target: %s:%s", seed, security.Ticker, security.Exchange)
					continue
				}
				relatedExchangeInfo, err = models.GetExchangeByTitle(database.DB, relatedExchange)
				if err != nil {
					log.Warnf("invalid exchange by title: %s for seed: %s - target: %s:%s", relatedExchange, seed, security.Ticker, security.Exchange)
					continue
				}
			} else {
				relatedExchangeInfo, err = models.GetExchangeBySuffixorPrefix(database.DB, relatedExchange, relatedExchange)
				if err != nil {
					log.Warnf("invalid exchange: %s for seed: %s - target: %s:%s", relatedExchange, seed, security.Ticker, security.Exchange)
					continue
				}
			}

			if !models.SecurityExists(database.DB, relatedTicker, relatedExchangeInfo.Title) {
				err = Scrape(relatedTicker, &relatedExchangeInfo.Title, manager, nil, nil, discoverer)
				if err != nil {
					log.Errorf("error scraping security(%s) related to %s: %v", relatedTicker+":"+relatedExchangeInfo.Title, security.Ticker+":"+security.Exchange, err)
					continue
				}
			}

			gap, err := models.GetSecurityTargetGapPercentage(database.DB, relatedTicker+":"+relatedExchangeInfo.Title)
			if err != nil {
				log.Warnf("error getting gap for %s:%s - target: %s:%s -> %v", relatedTicker, relatedExchangeInfo.Title, security.Ticker, security.Exchange, err)
				continue
			}

			log.Debugf("Gap for %s:%s -> %.2f", relatedTicker, relatedExchangeInfo.Title, gap)

			if gap > 0 {
				gapSums = append(gapSums, (float64(scrapedAllocation)/100)*(gap/100))
			}

			etf.RelatedSecurities = append(etf.RelatedSecurities, fmt.Sprintf("%s:%s:%d", relatedTicker, relatedExchangeInfo.Title, scrapedAllocation))

		}

		log.Debugf("Related securities for %s:%s -> %v", security.Ticker, security.Exchange, etf.RelatedSecurities)

		etf.Holdings = len(etf.RelatedSecurities)

		log.Debugf("GapSums  for %s:%s -> %v", security.Ticker, security.Exchange, gapSums)

		var increaseToTarget float64
		for _, gapSum := range gapSums {
			increaseToTarget += gapSum
		}

		log.Debugf("Increase to target for %s:%s -> %.2f", security.Ticker, security.Exchange, increaseToTarget)

		if increaseToTarget > 0 {
			etf.Target = models.NullableInt{
				Valid: true,
				Int64: int64(math.Floor(float64(security.Price) * (1 + (increaseToTarget / 100)))),
			}

			log.Debugf("Target for ETF %s:%s -> %d", security.Ticker, security.Exchange, etf.Target.Int64)
		}

		// if scrapedSeekingAlphaData.Holdings != nil {
		// 	etf.Holdings = *scrapedSeekingAlphaData.Holdings
		// }

		// Check if security already exists
		exists := models.SecurityExists(database.DB, security.Ticker, security.Exchange)
		if !exists {
			start := time.Now()
			err = models.CreateETF(database.DB, &etf)
			if err != nil {
				return fmt.Errorf("error creating ETF for seed (%s): %v", seed, err)
			}
			log.Infof("Created ETF based on Scraped data: %v", etf.PrettyPrintString())
			helpers.RecordDBQueryLatency("create_etf", start)
			helpers.RecordBusinessEvent("etf_created")
			helpers.RecordBusinessEvent("security_created")
		} else {
			start := time.Now()
			err = models.UpdateETF(database.DB, &etf)
			if err != nil {
				return fmt.Errorf("error updating ETF for seed (%s): %v", seed, err)
			}
			log.Infof("Updated ETF based on Scraped data: %v", etf.PrettyPrintString())
			helpers.RecordDBQueryLatency("update_etf", start)
			helpers.RecordBusinessEvent("etf_updated")
			helpers.RecordBusinessEvent("security_updated")
		}

	case "REIT":
		var reit models.REIT
		reit.Security = security
		// if scrapedSeekingAlphaData.FFO != nil {
		// 	reit.FFO = models.NullableInt{
		// 		Int64: int64(*scrapedSeekingAlphaData.FFO),
		// 		Valid: true,
		// 	}
		// }

		// if scrapedSeekingAlphaData.PFFO != nil {
		// 	reit.PFFO = models.NullableInt{
		// 		Int64: int64(*scrapedSeekingAlphaData.PFFO),
		// 		Valid: true,
		// 	}
		// }

		// if scrapedSeekingAlphaData.REITiming != nil {
		// 	reit.Timing = models.NullableString{
		// 		String: *scrapedSeekingAlphaData.REITiming,
		// 		Valid:  true,
		// 	}
		// }

		// Check if security already exists
		exists := models.SecurityExists(database.DB, security.Ticker, security.Exchange)
		if !exists {
			start := time.Now()
			err = models.CreateReit(database.DB, &reit)
			if err != nil {
				return fmt.Errorf("error creating REIT for seed (%s): %v", seed, err)
			}
			log.Infof("Created REIT based on Scraped data: %v", reit)
			helpers.RecordDBQueryLatency("create_reit", start)
			helpers.RecordBusinessEvent("reit_created")
			helpers.RecordBusinessEvent("security_created")
		} else {
			start := time.Now()
			err = models.UpdateREIT(database.DB, &reit)
			if err != nil {
				return fmt.Errorf("error updating REIT for seed (%s): %v", seed, err)
			}
			log.Infof("Updated REIT based on Scraped data: %v", reit)
			helpers.RecordDBQueryLatency("update_reit", start)
			helpers.RecordBusinessEvent("reit_updated")
			helpers.RecordBusinessEvent("security_updated")
		}
	default:
		return fmt.Errorf("invalid typology: %s - target: %s:%s", security.Typology, security.Ticker, security.Exchange)
	}

	cancel()

	return nil
}

func scrapeDividend(ticker string, exchange string, typology string, page *rod.Page) *models.Dividend {
	//Scrape Dividend Info if any
	var dividend models.Dividend
	dividend.Ticker = ticker
	dividend.Exchange = exchange

	var yieldStr string
	if typology == "ETF" {
		yieldStrElem, err := page.Timeout(5 * time.Second).Element(YH_YIELD_SELECTOR)
		if err != nil {
			log.Warnf("yield not found in page - target: %s:%s", ticker, exchange)
			return nil
		} else {
			yieldStr = yieldStrElem.MustText()
			log.Debugf("Scraped yield: %s", yieldStr)
			dividend.Timing = models.NullableString{
				String: string(models.TimingTTM),
				Valid:  true,
			}
		}

	} else {
		yieldStrElem, err := page.Timeout(5 * time.Second).Element(YH_FWD_YIELD_SELECTOR)
		if err != nil {
			log.Warnf("forward dividend & yield not found in page - target: %s:%s", ticker, exchange)
			return nil
		} else {
			yieldStr = yieldStrElem.MustText()
			yieldStr = extractPercentage(yieldStr)
			log.Debugf("Scraped forward dividend & yield: %s", yieldStr)
			dividend.Timing = models.NullableString{
				String: string(models.TimingFWD),
				Valid:  true,
			}
		}
	}

	yieldStr = helpers.NormalizeFloatStrToIntStr(yieldStr)
	if isAnEmptyString(yieldStr) {
		log.Warnf("empty yield: %s - target: %s:%s", yieldStr, ticker, exchange)
		return nil
	}

	scrapedYield, err := strconv.Atoi(yieldStr)
	if err != nil {
		log.Warnf("invalid yield: %s - target: %s:%s", yieldStr, ticker, exchange)
		return nil
	}

	if scrapedYield <= 0 {
		log.Warnf("invalid negative yield: %d - target: %s:%s", scrapedYield, ticker, exchange)
		return nil
	}

	dividend.Yield = scrapedYield

	return &dividend
}

func findExchangeInPage(ticker string, scrapingUrl string, browser *rod.Browser) (string, error) {
	log.Debugf("Scraping %s looking for exchange on url: %s", ticker, scrapingUrl)

	page, err := browser.Timeout(10 * time.Second).Page(proto.TargetCreateTarget{URL: scrapingUrl})
	if err != nil {
		return "", fmt.Errorf("error creating page: %s", err)
	}

	defer page.MustClose()

	exchangeElem, err := page.Timeout(5 * time.Second).Element(YH_EXCHANGE_SELECTOR)
	if err != nil {
		return "", fmt.Errorf("exchange not found in page - target: %s", ticker)
	} else {
		exchange := exchangeElem.MustText()
		if exchange == "" {
			return "", fmt.Errorf("empty exchange - target: %s", ticker)
		}
		log.Debugf("Scraped exchange for related security to etf: %s", exchange)

		exchange = strings.ToUpper(exchange)

		if !strings.Contains(exchange, "NYSE") && !strings.Contains(exchange, "NASDAQ") {
			return "", fmt.Errorf("invalid exchange: %s - target: %s", exchange, ticker)
		}

		if strings.Contains(exchange, "NYSE") {
			exchange = "NYSE"
		}

		if strings.Contains(exchange, "NASDAQ") {
			exchange = "NASDAQ"
		}

		if strings.Contains(exchange, "CBOE US") {
			exchange = "CBOEUS"
		}

		return exchange, nil
	}
}
