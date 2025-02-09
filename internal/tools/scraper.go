package tools

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Francesco99975/finexo/internal/database"
	"github.com/Francesco99975/finexo/internal/helpers"
	"github.com/Francesco99975/finexo/internal/models"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
	"github.com/labstack/gommon/log"
)

const BASE_SEEKINGALPHA_URL = "https://seekingalpha.com/symbol/" // TICKER:COUNTRY
const BASE_YAHOO_URL = "https://finance.yahoo.com/quote/"        // TICKER.EXCHANGE_SUFFIX
const BASE_MARKETBEAT_URL = "https://www.marketbeat.com/stocks/" // EXCHANGE_PREFIX/TICKER

func tickerExtractor(seed string) (string, string, error) {
	if len(seed) <= 0 {
		return "", "", fmt.Errorf("seed is empty")
	}

	seed = strings.ToUpper(seed)
	seed = strings.TrimSpace(seed)

	if strings.Contains(seed, ":") {
		parts := strings.Split(seed, ":")
		if len(parts) == 2 {
			return parts[1], parts[0], nil
		}
	} else if strings.Contains(seed, ".") {
		parts := strings.Split(seed, ".")
		if len(parts) == 2 {
			return parts[0], parts[1], nil
		}
	}
	return seed, "", nil

}

func isAnEmptyString(s string) bool {
	s = strings.TrimSpace(s)
	return s == "" || s == "N/A" || s == "-" || s == "--" || s == "n/a"
}

func Scrape(seed string, explicit_exchange *string) error {
	var security models.Security
	ticker, exchange_hint, err := tickerExtractor(seed)
	if err != nil {
		return fmt.Errorf("failed to extract ticker and exchange: %w", err)
	}

	security.Ticker = ticker
	var exchange *models.Exchange
	if exchange_hint != "" {
		exchange, err = models.GetExchangeBySuffixorPrefix(database.DB, exchange_hint, exchange_hint)
		if err != nil {
			return fmt.Errorf("failed to get exchange: %w", err)
		}
		security.Exchange = exchange.Title
	} else {
		if explicit_exchange != nil {
			exchange, err = models.GetExchangeByTitle(database.DB, *explicit_exchange)
			if err != nil {
				return fmt.Errorf("failed to get exchange: %w", err)
			}
			security.Exchange = exchange.Title
		} else {
			return fmt.Errorf("failed to extract exchange")
		}
	}

	// Run Rod in headless mode
	u := launcher.New().Headless(true).MustLaunch()
	browser := rod.New().ControlURL(u).MustConnect()
	defer browser.MustClose()

	log.Debugf("Scraping %s:%s", security.Ticker, security.Exchange)

	var page *rod.Page

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

	seekingalphaScrapingUrl := BASE_SEEKINGALPHA_URL + fmt.Sprintf("%s:%s", security.Ticker, exchange.CC)

	log.Debugf("Current scraping url: %s", yahooScrapingUrl)

	page, err = browser.Page(proto.TargetCreateTarget{URL: marketbeatScrapingUrl})
	if err != nil {
		log.Warnf("failed to open page on MarketBeat: %w. For seed %s", err, seed)
	}

	err = page.WaitLoad()
	if err != nil {
		log.Warnf("failed to wait for page load on MarketBeat: %w. For seed %s", err, seed)
	}

	//Scrape MarketBeat
	scrapedMarketBeatDataKeys, uperr := page.Timeout(5 * time.Second).Elements(".price-data-area dt")
	scrapedMarketBeatDataValues, err := page.Timeout(5 * time.Second).Elements(".price-data-area strong")
	if err != nil || uperr != nil {
		log.Warnf("failed to scrape MarketBeat data: %w. For seed %s", err, seed)
	} else {
		scrapedMarketBeatDataKeysArray := helpers.MapSlice(scrapedMarketBeatDataKeys, func(e *rod.Element) string {
			return e.MustText()
		})

		scrapedMarketBeatDataValuesArray := helpers.MapSlice(scrapedMarketBeatDataValues, func(e *rod.Element) string {
			return e.MustText()
		})

		for i := 0; i < len(scrapedMarketBeatDataKeysArray); i++ {
			key := strings.ToLower(scrapedMarketBeatDataKeysArray[i])

			if strings.Contains(key, "sector") {
				security.Sector = sql.NullString{String: scrapedMarketBeatDataValuesArray[i], Valid: true}
			}

			if strings.Contains(key, "industry") {
				security.Industry = sql.NullString{String: scrapedMarketBeatDataValuesArray[i], Valid: true}
			}

			if strings.Contains(key, "sub") {
				security.SubIndustry = sql.NullString{String: scrapedMarketBeatDataValuesArray[i], Valid: true}
			}

			if strings.Contains(key, "consensus") {
				security.Consensus = sql.NullString{String: scrapedMarketBeatDataValuesArray[i], Valid: true}
			}

			if strings.Contains(key, "score") {
				scrapedScoreStr := scrapedMarketBeatDataValuesArray[i]
				scrapedScoreStr = helpers.NormalizeFloatStrToIntStr(scrapedScoreStr)

				scrapedScore, err := strconv.Atoi(scrapedScoreStr)
				if err != nil {
					log.Warnf("failed to parse score: %w. For seed %s", err, seed)
				} else {
					security.Score = sql.NullInt64{Int64: int64(scrapedScore), Valid: true}
				}

			}

			if strings.Contains(key, "coverage") {
				scrapedCoverageStr := strings.Split(scrapedMarketBeatDataValuesArray[i], " ")[0]

				scrapedCoverage, err := strconv.Atoi(scrapedCoverageStr)
				if err != nil {
					log.Warnf("failed to parse coverage: %w. For seed %s", err, seed)
				} else {
					security.Coverage = sql.NullInt64{Int64: int64(scrapedCoverage), Valid: true}
				}
			}

			if strings.Contains(key, "outstanding") {
				scrapedOutstandingStr := scrapedMarketBeatDataValuesArray[i]
				scrapedOutstandingStr = helpers.NormalizeFloatStrToIntStr(scrapedOutstandingStr)

				scrapedOutstanding, err := strconv.ParseInt(scrapedOutstandingStr, 10, 64)
				if err != nil {
					log.Warnf("failed to parse outstanding: %w. For seed %s", err, seed)
				} else {
					security.Outstanding = sql.NullInt64{Int64: scrapedOutstanding, Valid: true}
				}
			}

		}
	}

	err = page.Navigate(seekingalphaScrapingUrl)
	if err != nil {
		log.Warnf("failed to open page on SeekingAlpha: %w. For seed %s", err, seed)
	}

	err = page.WaitLoad()
	if err != nil {
		log.Warnf("failed to wait for page load on SeekingAlpha: %w. For seed %s", err, seed)
	}

	//Scrape SeekingAlpha

	err = page.Navigate(yahooScrapingUrl)
	if err != nil {
		return fmt.Errorf("failed to open page on Yahoo: %w. For seed %s", err, seed)
	}

	err = page.WaitLoad()
	if err != nil {
		return fmt.Errorf("failed to wait for page load on Yahoo: %w. For seed %s", err, seed)
	}

	scrapedCurrencyElem, err := page.Timeout(5 * time.Second).Element("span.exchange.yf-wk4yba span:nth-child(3)")
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
	log.Debug("Scraped currency")

	scrapedFullNameElem, err := page.Timeout(5 * time.Second).Element(".yf-xxbei9")
	if err != nil {
		return fmt.Errorf("full name not found in page - target: %s:%s", security.Ticker, security.Exchange)
	}

	scrapedFullName := scrapedFullNameElem.MustText()
	scrapedFullName = strings.Split(scrapedFullName, " (")[0]

	if isAnEmptyString(scrapedFullName) {
		return fmt.Errorf("empty full name: %s - target: %s:%s", scrapedFullName, security.Ticker, security.Exchange)
	}

	security.FullName = scrapedFullName
	log.Debug("Scraped full name")

	scrapedTypology := "STOCK"
	if strings.Contains(strings.ToLower(scrapedFullName), "etf") || strings.Contains(strings.ToLower(scrapedFullName), "trust") {
		scrapedTypology = "ETF"
	} else if strings.Contains(strings.ToLower(scrapedFullName), "reit") {
		scrapedTypology = "REIT"
	}

	security.Typology = scrapedTypology
	log.Debugf("Scraped typology: %s", scrapedTypology)

	priceStrElem, err := page.Timeout(5 * time.Second).Element("span[data-testid='qsp-price']")
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

	priceChangeStrElem, err := page.Timeout(5 * time.Second).Element("span[data-testid='qsp-price-change']")
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

	priceChangePercentageStrElem, err := page.Timeout(5 * time.Second).Element("span[data-testid='qsp-price-change-percent']")
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
	log.Debug("Scraped price change percentage")

	yearlyRangeStrElem, err := page.Timeout(5 * time.Second).Element("[data-field='fiftyTwoWeekRange']")
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

	daylyRangeStrElem, err := page.Timeout(5 * time.Second).Element("[data-field='regularMarketDayRange']")
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

	marketCapStrElem, err := page.Timeout(5 * time.Second).Element("[data-field='marketCap']")
	if err != nil {
		log.Warnf("market cap not found in page - target: %s:%s", security.Ticker, security.Exchange)
		security.MarketCap = sql.NullInt64{
			Valid: false,
		}
	} else {
		marketCapStr := marketCapStrElem.MustText()
		log.Debugf("Scraped market cap: %s", marketCapStr)

		if isAnEmptyString(marketCapStr) {
			log.Warnf("empty market cap: %s - target: %s:%s", marketCapStr, security.Ticker, security.Exchange)
			security.MarketCap = sql.NullInt64{
				Valid: false,
			}
		} else {
			scrapedMarketCap, err := helpers.ParseNumberString(marketCapStr)
			if err != nil || scrapedMarketCap <= 0 {
				log.Warnf("invalid market cap: %s - target: %s:%s", marketCapStr, security.Ticker, security.Exchange)
				security.MarketCap = sql.NullInt64{
					Valid: false,
				}
			} else {
				security.MarketCap = sql.NullInt64{
					Int64: scrapedMarketCap,
					Valid: true,
				}
			}
		}
	}

	volumeStrElem, err := page.Timeout(5 * time.Second).Element("[data-field='regularMarketVolume']")
	if err != nil {
		log.Warnf("volume not found in page - target: %s:%s", security.Ticker, security.Exchange)
		security.Volume = sql.NullInt64{
			Valid: false,
		}
	} else {
		volumeStr := volumeStrElem.MustText()
		log.Debugf("Scraped volume: %s", volumeStr)
		volumeStr = strings.ReplaceAll(volumeStr, ",", "")

		if isAnEmptyString(volumeStr) {
			log.Warnf("empty volume: %s - target: %s:%s", volumeStr, security.Ticker, security.Exchange)
			security.Volume = sql.NullInt64{
				Valid: false,
			}
		} else {
			scrapedVolume, err := strconv.ParseInt(volumeStr, 10, 64)
			if err != nil || scrapedVolume <= 0 {
				log.Warnf("invalid volume: %s - target: %s:%s", volumeStr, security.Ticker, security.Exchange)
				security.Volume = sql.NullInt64{
					Valid: false,
				}
			} else {
				security.Volume = sql.NullInt64{
					Int64: scrapedVolume,
					Valid: true,
				}
			}
		}
	}

	log.Debug("Scraped volume")

	avgVolumeStrElem, err := page.Timeout(5 * time.Second).Element("[data-field='averageVolume']")
	if err != nil {
		log.Warnf("average volume not found in page - target: %s:%s", security.Ticker, security.Exchange)
		security.AvgVolume = sql.NullInt64{
			Valid: false,
		}
	} else {
		avgVolumeStr := avgVolumeStrElem.MustText()
		log.Debugf("Scraped average volume: %s", avgVolumeStr)
		avgVolumeStr = strings.ReplaceAll(avgVolumeStr, ",", "")

		if isAnEmptyString(avgVolumeStr) {
			log.Warnf("empty average volume: %s - target: %s:%s", avgVolumeStr, security.Ticker, security.Exchange)
			security.AvgVolume = sql.NullInt64{
				Valid: false,
			}
		} else {
			scrapedAvgVolume, err := strconv.ParseInt(avgVolumeStr, 10, 64)
			if err != nil || scrapedAvgVolume <= 0 {
				log.Warnf("invalid average volume: %s - target: %s:%s", avgVolumeStr, security.Ticker, security.Exchange)
				security.AvgVolume = sql.NullInt64{
					Valid: false,
				}
			} else {
				security.AvgVolume = sql.NullInt64{
					Int64: scrapedAvgVolume,
					Valid: true,
				}
			}
		}

	}

	log.Debug("Scraped average volume")

	betaStrElem, err := page.Timeout(5 * time.Second).Element("span[title='Beta (5Y Monthly)'] ~ span")
	if err != nil {
		log.Warnf("beta not found in page - target: %s:%s", security.Ticker, security.Exchange)
		security.Beta = sql.NullInt64{
			Valid: false,
		}
	} else {
		betaStr := betaStrElem.MustText()
		log.Debugf("Scraped beta: %s", betaStr)
		betaStr = helpers.NormalizeFloatStrToIntStr(betaStr)
		if isAnEmptyString(betaStr) {
			log.Warnf("empty beta: %s - target: %s:%s", betaStr, security.Ticker, security.Exchange)
			security.Beta = sql.NullInt64{
				Valid: false,
			}
		} else {
			scrapedBeta, err := strconv.Atoi(betaStr)
			if err != nil || scrapedBeta <= 0 {
				log.Warnf("invalid beta: %s - target: %s:%s", betaStr, security.Ticker, security.Exchange)
				security.Beta = sql.NullInt64{
					Valid: false,
				}
			} else {
				security.Beta = sql.NullInt64{
					Int64: int64(scrapedBeta),
					Valid: true,
				}
			}
		}
	}

	log.Debug("Scraped beta")

	pcloseStrElem, err := page.Timeout(5 * time.Second).Element("[data-field='regularMarketPreviousClose']")
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

	copenStrElem, err := page.Timeout(5 * time.Second).Element("[data-field='regularMarketOpen']")
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

	bidPayloadStrElem, err := page.Timeout(5 * time.Second).Element("span[title='Bid'] ~ span")
	if err != nil {
		return fmt.Errorf("bid not found in page - target: %s:%s", security.Ticker, security.Exchange)
	}

	bidPayloadStr := bidPayloadStrElem.MustText()
	log.Debugf("Scraped bid: %s", bidPayloadStr)
	bidPayloadStr = strings.ReplaceAll(bidPayloadStr, " ", "")
	bidPayloadArr := strings.Split(bidPayloadStr, "x")
	if len(bidPayloadArr) != 2 {
		return fmt.Errorf("invalid bid payload: %s - target: %s:%s", bidPayloadStr, security.Ticker, security.Exchange)
	}

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
		security.BidSize = sql.NullInt64{
			Valid: false,
		}
	} else {
		scrapedBidSize, err := strconv.Atoi(bidSizeStr)
		if err != nil || scrapedBidSize < 0 {
			log.Warnf("invalid bid size: %s - target: %s:%s", bidSizeStr, security.Ticker, security.Exchange)
			security.BidSize = sql.NullInt64{
				Valid: false,
			}
		} else {
			security.BidSize = sql.NullInt64{
				Int64: int64(scrapedBidSize),
				Valid: true,
			}
		}

	}
	log.Debug("Scraped bid size")

	askPayloadStrElem, err := page.Timeout(5 * time.Second).Element("span[title='Ask'] ~ span")
	if err != nil {
		return fmt.Errorf("ask not found in page - target: %s:%s", security.Ticker, security.Exchange)
	}

	askPayloadStr := askPayloadStrElem.MustText()
	log.Debugf("Scraped ask: %s", askPayloadStr)
	askPayloadStr = strings.ReplaceAll(askPayloadStr, " ", "")
	askPayloadArr := strings.Split(askPayloadStr, "x")
	if len(askPayloadArr) != 2 {
		return fmt.Errorf("invalid ask payload: %s - target: %s:%s", askPayloadStr, security.Ticker, security.Exchange)
	}

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
		security.AskSize = sql.NullInt64{
			Valid: false,
		}
	} else {
		scrapedAskSize, err := strconv.Atoi(askSizeStr)
		if err != nil || scrapedAskSize <= 0 {
			log.Warnf("invalid ask size: %s - target: %s:%s", askSizeStr, security.Ticker, security.Exchange)
			security.AskSize = sql.NullInt64{
				Valid: false,
			}
		}

		security.AskSize = sql.NullInt64{
			Int64: int64(scrapedAskSize),
			Valid: true,
		}
	}

	log.Debug("Scraped ask size")

	stockDataElements, err := page.Timeout(5 * time.Second).Elements("[data-field='trailingPE']")
	if err != nil {
		return fmt.Errorf("trailing PE not found in page - target: %s:%s", security.Ticker, security.Exchange)
	}

	if len(stockDataElements) == 0 {
		log.Warnf("empty trailing PE: %s - target: %s:%s", stockDataElements, security.Ticker, security.Exchange)
		security.PE = sql.NullInt64{
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
			security.PE = sql.NullInt64{
				Valid: false,
			}
		} else {
			security.PE = sql.NullInt64{
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
		if err != nil || scrapedEps <= 0 {
			log.Warnf("invalid EPS: %s - target: %s:%s", epsStr, security.Ticker, security.Exchange)
			security.EPS = sql.NullInt64{
				Valid: false,
			}
		} else {
			security.EPS = sql.NullInt64{
				Int64: int64(scrapedEps),
				Valid: true,
			}
		}
	}

	log.Debug("Scraped EPS")

	security.STM = sql.NullString{
		String: string(models.TimingTTM),
		Valid:  true,
	}

	security.Dividend = scrapeDividend(ticker, security.Exchange, security.Typology, page)
	log.Debug("Scraped dividend")

	switch security.Typology {
	case "STOCK":
		// err = models.CreateStock(database.DB, &security)
		// if err != nil {
		// 	return err
		// }
		log.Infof("Scraped data: %v", security)
	case "ETF":
		var etf models.ETF

		etf.Security = security

		aumStrElem, err := page.Timeout(5 * time.Second).Element("span[title='Net Assets'] ~ span")
		if err != nil {
			log.Warnf("AUM not found in page - target: %s:%s", security.Ticker, security.Exchange)
			etf.AUM = sql.NullInt64{
				Valid: false,
			}
		} else {
			aumStr := aumStrElem.MustText()
			log.Debugf("Scraped AUM: %s", aumStr)
			scrapedAum, err := helpers.ParseNumberString(aumStr)
			if err != nil || scrapedAum <= 0 {
				log.Warnf("invalid AUM: %s - target: %s:%s", aumStr, security.Ticker, security.Exchange)
				etf.AUM = sql.NullInt64{
					Valid: false,
				}
			} else {
				etf.AUM = sql.NullInt64{
					Int64: scrapedAum,
					Valid: true,
				}
			}
		}

		log.Debug("Scraped AUM")

		erStrElem, err := page.Timeout(5 * time.Second).Element("span[title='Expense Ratio (net)'] ~ span")
		if err != nil {
			log.Warnf("expense ratio not found in page - target: %s:%s", security.Ticker, security.Exchange)
			etf.ExpenseRatio = sql.NullInt64{
				Valid: false,
			}
		} else {
			erStr := erStrElem.MustText()
			log.Debugf("Scraped expense ratio: %s", erStr)
			erStr = helpers.NormalizeFloatStrToIntStr(erStr)
			if erStr == "" {
				log.Warnf("empty expense ratio: %s - target: %s:%s", erStr, security.Ticker, security.Exchange)
				etf.ExpenseRatio = sql.NullInt64{
					Valid: false,
				}
			} else {
				scrapedEr, err := strconv.Atoi(erStr)
				if err != nil || scrapedEr <= 0 {
					log.Warnf("invalid expense ratio: %s - target: %s:%s", erStr, security.Ticker, security.Exchange)
					etf.ExpenseRatio = sql.NullInt64{
						Valid: false,
					}
				} else {
					etf.ExpenseRatio = sql.NullInt64{
						Int64: int64(scrapedEr),
						Valid: true,
					}
				}
			}
		}

		log.Debug("Scraped expense ratio")

		navStrElem, err := page.Timeout(5 * time.Second).Element("span[title='NAV'] ~ span")
		if err != nil {
			log.Warnf("NAV not found in page - target: %s:%s", security.Ticker, security.Exchange)
			etf.NAV = sql.NullInt64{
				Valid: false,
			}
		} else {
			navStr := navStrElem.MustText()
			log.Debugf("Scraped NAV: %s", navStr)
			navStr = helpers.NormalizeFloatStrToIntStr(navStr)
			if isAnEmptyString(navStr) {
				log.Warnf("empty NAV: %s - target: %s:%s", navStr, security.Ticker, security.Exchange)
				etf.NAV = sql.NullInt64{
					Valid: false,
				}
			} else {
				scrapedNav, err := strconv.Atoi(navStr)
				if err != nil || scrapedNav <= 0 {
					log.Warnf("invalid NAV: %s - target: %s:%s", navStr, security.Ticker, security.Exchange)
					etf.NAV = sql.NullInt64{
						Valid: false,
					}
				} else {
					etf.NAV = sql.NullInt64{
						Int64: int64(scrapedNav),
						Valid: true,
					}
				}
			}
		}

		log.Debug("Scraped NAV")

		inceptionDateStrElem, err := page.Timeout(5 * time.Second).Elements("section[data-testid='company-overview-card'] p[title]")
		if err != nil {
			log.Warnf("inception date not found in page - target: %s:%s", security.Ticker, security.Exchange)
			etf.InceptionDate = sql.NullTime{
				Valid: false,
			}
		} else {
			inceptionDateStr := inceptionDateStrElem[3].MustText()
			log.Debugf("Scraped inception date: %s", inceptionDateStr)
			if isAnEmptyString(inceptionDateStr) {
				log.Warnf("empty inception date: %s - target: %s:%s", inceptionDateStr, security.Ticker, security.Exchange)
				etf.InceptionDate = sql.NullTime{
					Valid: false,
				}
			} else {
				scrapedInceptionDate, err := time.Parse("2006-01-02", inceptionDateStr)
				if err != nil {
					log.Warnf("invalid inception date: %s - target: %s:%s", inceptionDateStr, security.Ticker, security.Exchange)
					etf.InceptionDate = sql.NullTime{
						Valid: false,
					}
				} else {
					etf.InceptionDate = sql.NullTime{
						Time:  scrapedInceptionDate,
						Valid: true,
					}
				}
			}
		}

		log.Debug("Scraped inception date")

		relationsElementsTickers, err := page.Timeout(5 * time.Second).Elements("section[data-testid='top-holdings'] a[data-testid='ticker-container']")
		if err != nil {
			log.Warnf("top holdings not found in page - target: %s:%s", security.Ticker, security.Exchange)
		}

		relationsElementsAllocations, err := page.Timeout(5 * time.Second).Elements("section[data-testid='top-holdings'] a[data-testid='ticker-container'] ~ span.data")
		if err != nil {
			log.Warnf("top holdings not found in page - target: %s:%s", security.Ticker, security.Exchange)
		}

		for i := range len(relationsElementsTickers) {
			seed := relationsElementsTickers[i].MustText()
			log.Debugf("Scraped top holding: %s", security.Ticker)
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
			allocationStr := relationsElementsAllocations[i].MustText()
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
			if relatedExchange == "" {
				relatedExchange, err = findExchangeInPage(ticker)
				if err != nil {
					log.Warnf("invalid exchange or could not find: %s - target: %s:%s", seed, security.Ticker, security.Exchange)
					continue
				}
			}
			relatedExchangeInfo, err := models.GetExchangeBySuffixorPrefix(database.DB, relatedExchange, relatedExchange)
			if err != nil {
				log.Warnf("invalid exchange: %s - target: %s:%s", relatedExchange, security.Ticker, security.Exchange)
				continue
			}

			if !models.SecurityExists(database.DB, relatedTicker, relatedExchangeInfo.Title) {
				err := Scrape(relatedTicker, &relatedExchangeInfo.Title)
				if err != nil {
					log.Errorf("error scraping related security: %s", err)
					continue
				}
			}

			etf.RelatedSecurities = append(etf.RelatedSecurities, fmt.Sprintf("%s:%s:%d", relatedTicker, relatedExchange, scrapedAllocation))
		}

		//Display In a good format all the scraped data
		log.Infof("Scraped data: %v", etf)

		// err = models.CreateETF(database.DB, &etf)
		// if err != nil {
		// 	return err
		// }

	case "REIT":
		security.STM = sql.NullString{
			String: string(models.TimingTTM),
			Valid:  true,
		}
	default:
		return fmt.Errorf("invalid typology: %s - target: %s:%s", security.Typology, security.Ticker, security.Exchange)
	}

	page.MustClose()

	return nil
}

func scrapeDividend(ticker string, exchange string, typology string, page *rod.Page) *models.Dividend {
	//Scrape Dividend Info if any
	var dividend models.Dividend

	var yieldStr string
	if typology == "ETF" {
		yieldStrElem, err := page.Timeout(5 * time.Second).Element("span[title='Yield'] ~ span")
		if err != nil {
			log.Warnf("yield not found in page - target: %s:%s", ticker, exchange)
			return nil
		} else {
			yieldStr = yieldStrElem.MustText()
			log.Debugf("Scraped yield: %s", yieldStr)
		}
	} else {
		yieldStrElem, err := page.Timeout(5 * time.Second).Element("span[title='Forward Dividend & Yield'] ~ span")
		if err != nil {
			log.Warnf("forward dividend & yield not found in page - target: %s:%s", ticker, exchange)
			return nil
		} else {
			yieldStr = yieldStrElem.MustText()
			log.Debugf("Scraped forward dividend & yield: %s", yieldStr)
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

func findExchangeInPage(ticker string) (string, error) {
	// Run Rod in headless mode
	u := launcher.New().Headless(true).MustLaunch()
	browser := rod.New().ControlURL(u).MustConnect()

	log.Debugf("Scraping %s looking for exchange", ticker)

	var page *rod.Page
	exchangeElem, err := page.Timeout(5 * time.Second).Element("span.exchange span")
	if err != nil {
		return "", fmt.Errorf("exchange not found in page - target: %s", ticker)
	}
	exchange := exchangeElem.MustText()

	page.MustClose()
	browser.MustClose()

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

	return exchange, nil
}
