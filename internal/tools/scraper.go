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
	"github.com/labstack/gommon/log"
)

const BASE_SEEKINGALPHA_URL = "https://seekingalpha.com/symbol/" // TICKER:COUNTRY
const BASE_YAHOO_URL = "https://finance.yahoo.com/quote/"        // TICKER.EXCHANGE_SUFFIX
const BASE_MARKETBEAT_URL = "https://www.marketbeat.com/stocks/" // EXCHANGE_PREFIX/TICKER

func isAnEmptyString(s string) bool {
	s = strings.TrimSpace(s)
	return s == "" || s == "N/A" || s == "-" || s == "--" || s == "n/a"
}

func Scrape(ticker string, exchange models.Exchange, country models.Country) error {
	var security models.Security

	// Run Rod in headless mode
	u := launcher.New().Headless(true).MustLaunch()
	browser := rod.New().ControlURL(u).MustConnect()

	var page *rod.Page

	if exchange.Suffix.Valid {
		page = browser.MustPage(BASE_YAHOO_URL + fmt.Sprintf("%s.%s", ticker, exchange.Suffix.String)).MustWaitLoad()
	} else {
		page = browser.MustPage(BASE_YAHOO_URL + ticker).MustWaitLoad()
	}

	scrapedCurrencyElem, err := page.Element("span.exchange.yf-wk4yba span:nth-child(3)")
	if err != nil {
		return fmt.Errorf("currency not found in page - target: %s:%s", ticker, exchange.Title)
	}

	scrapedCurrency := scrapedCurrencyElem.MustText()

	scrapedCurrency = strings.TrimSpace(scrapedCurrency)

	if scrapedCurrency != country.Currency {
		return fmt.Errorf("currency mismatch: %s != %s - target: %s:%s", scrapedCurrency, country.Currency, ticker, exchange.Title)
	}

	scrapedFullNameElem, err := page.Element(".yf-xxbei9")
	if err != nil {
		return fmt.Errorf("full name not found in page - target: %s:%s", ticker, exchange.Title)
	}

	scrapedFullName := scrapedFullNameElem.MustText()
	scrapedFullName = strings.Split(scrapedFullName, " ")[0]

	if isAnEmptyString(scrapedFullName) {
		return fmt.Errorf("empty full name: %s - target: %s:%s", scrapedFullName, ticker, exchange.Title)
	}

	security.FullName = scrapedFullName

	scrapedTypology := "STOCK"
	if strings.Contains(scrapedFullName, "ETF") {
		scrapedTypology = "ETF"
	} else if strings.Contains(scrapedFullName, "REIT") {
		scrapedTypology = "REIT"
	}

	security.Typology = scrapedTypology

	priceStrElem, err := page.Element("span[data-testid='qsp-price']")
	if err != nil {
		return fmt.Errorf("price not found in page - target: %s:%s", ticker, exchange.Title)
	}

	priceStr := priceStrElem.MustText()
	priceStr = helpers.NormalizeFloatStrToIntStr(priceStr)

	if isAnEmptyString(priceStr) {
		return fmt.Errorf("empty price: %s - target: %s:%s", priceStr, ticker, exchange.Title)
	}

	scrapedPrice, err := strconv.Atoi(priceStr)
	if err != nil {
		return fmt.Errorf("invalid price: %s - target: %s:%s", priceStr, ticker, exchange.Title)
	}

	if scrapedPrice <= 0 {
		return fmt.Errorf("invalid negative price: %d - target: %s:%s", scrapedPrice, ticker, exchange.Title)
	}

	security.Price = scrapedPrice

	priceChangeStrElem, err := page.Element("span[data-testid='qsp-price-change']")
	if err != nil {
		return fmt.Errorf("price change not found in page - target: %s:%s", ticker, exchange.Title)
	}
	priceChangeStr := priceChangeStrElem.MustText()
	priceChangeStr = helpers.NormalizeFloatStrToIntStr(priceChangeStr)

	if isAnEmptyString(priceChangeStr) {
		return fmt.Errorf("empty price change: %s - target: %s:%s", priceChangeStr, ticker, exchange.Title)
	}

	scrapedPriceChange, err := strconv.Atoi(priceChangeStr)
	if err != nil {
		return fmt.Errorf("invalid price change: %s - target: %s:%s", priceChangeStr, ticker, exchange.Title)
	}

	security.PC = scrapedPriceChange

	priceChangePercentageStrElem, err := page.Element("span[data-testid='qsp-price-change-percentage']")
	if err != nil {
		return fmt.Errorf("price change percentage not found in page - target: %s:%s", ticker, exchange.Title)
	}

	priceChangePercentageStr := priceChangePercentageStrElem.MustText()
	priceChangePercentageStr = helpers.NormalizeFloatStrToIntStr(priceChangePercentageStr)

	if priceChangePercentageStr == "" {
		return fmt.Errorf("empty price change percentage: %s - target: %s:%s", priceChangePercentageStr, ticker, exchange.Title)
	}

	scrapedPriceChangePercentage, err := strconv.Atoi(priceChangePercentageStr)
	if err != nil {
		return fmt.Errorf("invalid price change percentage: %s - target: %s:%s", priceChangePercentageStr, ticker, exchange.Title)
	}

	security.PCP = scrapedPriceChangePercentage

	yearlyRangeStrElem, err := page.Element("[data-field='fiftyTwoWeekRange']")
	if err != nil {
		return fmt.Errorf("yearly range not found in page - target: %s:%s", ticker, exchange.Title)
	}

	yearlyRangeStr := yearlyRangeStrElem.MustText()

	yearlyRangeStr = strings.ReplaceAll(yearlyRangeStr, " ", "")
	yearlyRangeArr := strings.Split(yearlyRangeStr, "-")

	if len(yearlyRangeArr) != 2 {
		return fmt.Errorf("invalid yearly range: %s - target: %s:%s", yearlyRangeStr, ticker, exchange.Title)
	}

	yrlStr := yearlyRangeArr[0]
	yrlStr = helpers.NormalizeFloatStrToIntStr(yrlStr)
	if yrlStr == "" {
		return fmt.Errorf("empty yearly range low: %s - target: %s:%s", yrlStr, ticker, exchange.Title)
	}

	scrapedYrl, err := strconv.Atoi(yrlStr)
	if err != nil {
		return fmt.Errorf("invalid yearly range low: %s - target: %s:%s", yrlStr, ticker, exchange.Title)
	}

	if scrapedYrl <= 0 {
		return fmt.Errorf("invalid negative yearly range low: %d - target: %s:%s", scrapedYrl, ticker, exchange.Title)
	}

	security.YearLow = scrapedYrl

	yrhStr := yearlyRangeArr[1]
	yrhStr = helpers.NormalizeFloatStrToIntStr(yrhStr)
	if yrhStr == "" {
		return fmt.Errorf("empty yearly range high: %s - target: %s:%s", yrhStr, ticker, exchange.Title)
	}

	scrapedYrh, err := strconv.Atoi(yrhStr)
	if err != nil {
		return fmt.Errorf("invalid yearly range high: %s - target: %s:%s", yrhStr, ticker, exchange.Title)
	}

	if scrapedYrh <= 0 {
		return fmt.Errorf("invalid negative yearly range high: %d - target: %s:%s", scrapedYrh, ticker, exchange.Title)
	}

	if scrapedYrh < scrapedYrl {
		return fmt.Errorf("invalid yearly range high < low: %d < %d - target: %s:%s", scrapedYrh, scrapedYrl, ticker, exchange.Title)
	}

	security.YearHigh = scrapedYrh

	daylyRangeStrElem, err := page.Element("[data-field='regularMarketDayRange']")
	if err != nil {
		return fmt.Errorf("daily range not found in page - target: %s:%s", ticker, exchange.Title)
	}

	daylyRangeStr := daylyRangeStrElem.MustText()

	daylyRangeStr = strings.ReplaceAll(daylyRangeStr, " ", "")
	daylyRangeArr := strings.Split(daylyRangeStr, "-")
	if len(daylyRangeArr) != 2 {
		return fmt.Errorf("invalid daily range: %s - target: %s:%s", daylyRangeStr, ticker, exchange.Title)
	}

	drlStr := daylyRangeArr[0]
	drlStr = helpers.NormalizeFloatStrToIntStr(drlStr)
	if drlStr == "" {
		return fmt.Errorf("empty daily range low: %s - target: %s:%s", drlStr, ticker, exchange.Title)
	}

	scrapedDrl, err := strconv.Atoi(drlStr)
	if err != nil {
		return fmt.Errorf("invalid daily range low: %s - target: %s:%s", drlStr, ticker, exchange.Title)
	}

	if scrapedDrl <= 0 {
		return fmt.Errorf("invalid negative daily range low: %d - target: %s:%s", scrapedDrl, ticker, exchange.Title)
	}

	security.DayLow = scrapedDrl

	drhStr := daylyRangeArr[1]
	drhStr = helpers.NormalizeFloatStrToIntStr(drhStr)
	if drhStr == "" {
		return fmt.Errorf("empty daily range high: %s - target: %s:%s", drhStr, ticker, exchange.Title)
	}

	scrapedDrh, err := strconv.Atoi(drhStr)
	if err != nil {
		return fmt.Errorf("invalid daily range high: %s - target: %s:%s", drhStr, ticker, exchange.Title)
	}

	if scrapedDrh <= 0 {
		return fmt.Errorf("invalid negative daily range high: %d - target: %s:%s", scrapedDrh, ticker, exchange.Title)
	}

	if scrapedDrh < scrapedDrl {
		return fmt.Errorf("invalid daily range high < low: %d < %d - target: %s:%s", scrapedDrh, scrapedDrl, ticker, exchange.Title)
	}

	security.DayHigh = scrapedDrh

	marketCapStrElem, err := page.Element("[data-field='marketCap']")
	if err != nil {
		log.Errorf("market cap not found in page - target: %s:%s", ticker, exchange.Title)
		security.MarketCap = sql.NullInt64{
			Valid: false,
		}
	}

	marketCapStr := marketCapStrElem.MustText()

	if isAnEmptyString(marketCapStr) {
		log.Errorf("empty market cap: %s - target: %s:%s", marketCapStr, ticker, exchange.Title)
		security.MarketCap = sql.NullInt64{
			Valid: false,
		}
	} else {
		scrapedMarketCap, err := helpers.ParseNumberString(marketCapStr)
		if err != nil || scrapedMarketCap <= 0 {
			log.Errorf("invalid market cap: %s - target: %s:%s", marketCapStr, ticker, exchange.Title)
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

	volumeStrElem, err := page.Element("[data-field='regularMarketVolume']")
	if err != nil {
		log.Errorf("volume not found in page - target: %s:%s", ticker, exchange.Title)
		security.Volume = sql.NullInt64{
			Valid: false,
		}
	}

	volumeStr := volumeStrElem.MustText()
	volumeStr = strings.ReplaceAll(volumeStr, ",", "")

	if isAnEmptyString(volumeStr) {
		log.Errorf("empty volume: %s - target: %s:%s", volumeStr, ticker, exchange.Title)
		security.Volume = sql.NullInt64{
			Valid: false,
		}
	} else {
		scrapedVolume, err := strconv.ParseInt(volumeStr, 10, 64)
		if err != nil || scrapedVolume <= 0 {
			log.Errorf("invalid volume: %s - target: %s:%s", volumeStr, ticker, exchange.Title)
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

	avgVolumeStrElem, err := page.Element("[data-field='averageVolume']")
	if err != nil {
		log.Errorf("average volume not found in page - target: %s:%s", ticker, exchange.Title)
		security.AvgVolume = sql.NullInt64{
			Valid: false,
		}
	}
	avgVolumeStr := avgVolumeStrElem.MustText()
	avgVolumeStr = strings.ReplaceAll(avgVolumeStr, ",", "")

	if isAnEmptyString(avgVolumeStr) {
		log.Errorf("empty average volume: %s - target: %s:%s", avgVolumeStr, ticker, exchange.Title)
		security.AvgVolume = sql.NullInt64{
			Valid: false,
		}
	} else {
		scrapedAvgVolume, err := strconv.ParseInt(avgVolumeStr, 10, 64)
		if err != nil || scrapedAvgVolume <= 0 {
			log.Errorf("invalid average volume: %s - target: %s:%s", avgVolumeStr, ticker, exchange.Title)
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

	betaStrElem, err := page.Element("span[title='Beta (5Y Monthly)'] ~ span")
	if err != nil {
		log.Errorf("beta not found in page - target: %s:%s", ticker, exchange.Title)
		security.Beta = sql.NullInt64{
			Valid: false,
		}
	}
	betaStr := betaStrElem.MustText()
	betaStr = helpers.NormalizeFloatStrToIntStr(betaStr)
	if isAnEmptyString(betaStr) {
		log.Errorf("empty beta: %s - target: %s:%s", betaStr, ticker, exchange.Title)
		security.Beta = sql.NullInt64{
			Valid: false,
		}
	} else {
		scrapedBeta, err := strconv.Atoi(betaStr)
		if err != nil || scrapedBeta <= 0 {
			log.Errorf("invalid beta: %s - target: %s:%s", betaStr, ticker, exchange.Title)
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

	pcloseStrElem, err := page.Element("[data-field='regularMarketPreviousClose']")
	if err != nil {
		return fmt.Errorf("previous close not found in page - target: %s:%s", ticker, exchange.Title)
	}

	pcloseStr := pcloseStrElem.MustText()
	pcloseStr = helpers.NormalizeFloatStrToIntStr(pcloseStr)
	if isAnEmptyString(pcloseStr) {
		return fmt.Errorf("empty previous close: %s - target: %s:%s", pcloseStr, ticker, exchange.Title)
	}

	scrapedPclose, err := strconv.Atoi(pcloseStr)
	if err != nil {
		return fmt.Errorf("invalid previous close: %s - target: %s:%s", pcloseStr, ticker, exchange.Title)
	}

	if scrapedPclose <= 0 {
		return fmt.Errorf("invalid negative previous close: %d - target: %s:%s", scrapedPclose, ticker, exchange.Title)
	}

	security.PClose = scrapedPclose

	copenStrElem, err := page.Element("[data-field='regularMarketOpen']")
	if err != nil {
		return fmt.Errorf("open not found in page - target: %s:%s", ticker, exchange.Title)
	}
	copenStr := copenStrElem.MustText()
	copenStr = helpers.NormalizeFloatStrToIntStr(copenStr)
	if isAnEmptyString(copenStr) {
		return fmt.Errorf("empty open: %s - target: %s:%s", copenStr, ticker, exchange.Title)
	}

	scrapedCopen, err := strconv.Atoi(copenStr)
	if err != nil {
		return fmt.Errorf("invalid open: %s - target: %s:%s", copenStr, ticker, exchange.Title)
	}

	if scrapedCopen <= 0 {
		return fmt.Errorf("invalid negative open: %d - target: %s:%s", scrapedCopen, ticker, exchange.Title)
	}

	security.COpen = scrapedCopen

	bidPayloadStrElem, err := page.Element("span[title='Bid'] ~ span")
	if err != nil {
		return fmt.Errorf("bid not found in page - target: %s:%s", ticker, exchange.Title)
	}

	bidPayloadStr := bidPayloadStrElem.MustText()
	bidPayloadStr = strings.ReplaceAll(bidPayloadStr, " ", "")
	bidPayloadArr := strings.Split(bidPayloadStr, "x")
	if len(bidPayloadArr) != 2 {
		return fmt.Errorf("invalid bid payload: %s - target: %s:%s", bidPayloadStr, ticker, exchange.Title)
	}

	bidStr := bidPayloadArr[0]
	bidStr = helpers.NormalizeFloatStrToIntStr(bidStr)
	if bidStr == "" {
		return fmt.Errorf("empty bid: %s - target: %s:%s", bidStr, ticker, exchange.Title)
	}

	scrapedBid, err := strconv.Atoi(bidStr)
	if err != nil {
		return fmt.Errorf("invalid bid: %s - target: %s:%s", bidStr, ticker, exchange.Title)
	}

	if scrapedBid <= 0 {
		return fmt.Errorf("invalid negative bid: %d - target: %s:%s", scrapedBid, ticker, exchange.Title)
	}

	security.Bid = scrapedBid

	bidSizeStr := bidPayloadArr[1]
	if isAnEmptyString(bidSizeStr) {
		log.Errorf("empty bid size: %s - target: %s:%s", bidSizeStr, ticker, exchange.Title)
		security.BidSize = sql.NullInt64{
			Valid: false,
		}
	} else {
		scrapedBidSize, err := strconv.Atoi(bidSizeStr)
		if err != nil || scrapedBidSize < 0 {
			log.Errorf("invalid bid size: %s - target: %s:%s", bidSizeStr, ticker, exchange.Title)
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

	askPayloadStrElem, err := page.Element("span[title='Ask'] ~ span")
	if err != nil {
		return fmt.Errorf("ask not found in page - target: %s:%s", ticker, exchange.Title)
	}

	askPayloadStr := askPayloadStrElem.MustText()
	askPayloadStr = strings.ReplaceAll(askPayloadStr, " ", "")
	askPayloadArr := strings.Split(askPayloadStr, "x")
	if len(askPayloadArr) != 2 {
		return fmt.Errorf("invalid ask payload: %s - target: %s:%s", askPayloadStr, ticker, exchange.Title)
	}

	askStr := askPayloadArr[0]
	askStr = helpers.NormalizeFloatStrToIntStr(askStr)
	if isAnEmptyString(askStr) {
		return fmt.Errorf("empty ask: %s - target: %s:%s", askStr, ticker, exchange.Title)
	}

	scrapedAsk, err := strconv.Atoi(askStr)
	if err != nil {
		return fmt.Errorf("invalid ask: %s - target: %s:%s", askStr, ticker, exchange.Title)
	}

	if scrapedAsk <= 0 {
		return fmt.Errorf("invalid negative ask: %d - target: %s:%s", scrapedAsk, ticker, exchange.Title)
	}

	security.Ask = scrapedAsk

	askSizeStr := askPayloadArr[1]
	if isAnEmptyString(askSizeStr) {
		log.Errorf("empty ask size: %s - target: %s:%s", askSizeStr, ticker, exchange.Title)
		security.AskSize = sql.NullInt64{
			Valid: false,
		}
	} else {
		scrapedAskSize, err := strconv.Atoi(askSizeStr)
		if err != nil || scrapedAskSize <= 0 {
			log.Errorf("invalid ask size: %s - target: %s:%s", askSizeStr, ticker, exchange.Title)
			security.AskSize = sql.NullInt64{
				Valid: false,
			}
		}

		security.AskSize = sql.NullInt64{
			Int64: int64(scrapedAskSize),
			Valid: true,
		}
	}

	stockDataElements, err := page.Elements("[data-field='trailingPE']")
	if err != nil {
		return fmt.Errorf("trailing PE not found in page - target: %s:%s", ticker, exchange.Title)
	}

	if len(stockDataElements) == 0 {
		log.Errorf("empty trailing PE: %s - target: %s:%s", stockDataElements, ticker, exchange.Title)
		security.PE = sql.NullInt64{
			Valid: false,
		}
	}

	if len(stockDataElements) == 1 || len(stockDataElements) == 2 {
		peStr := stockDataElements[0].MustText()
		peStr = helpers.NormalizeFloatStrToIntStr(peStr)
		if peStr == "" {
			return fmt.Errorf("empty trailing PE: %s - target: %s:%s", peStr, ticker, exchange.Title)
		}

		scrapedPe, err := strconv.Atoi(peStr)
		if err != nil || scrapedPe <= 0 {
			log.Errorf("invalid trailing PE: %s - target: %s:%s", peStr, ticker, exchange.Title)
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

	if len(stockDataElements) == 2 {
		epsStr := stockDataElements[1].MustText()
		epsStr = helpers.NormalizeFloatStrToIntStr(epsStr)
		if epsStr == "" {
			return fmt.Errorf("empty EPS: %s - target: %s:%s", epsStr, ticker, exchange.Title)
		}

		scrapedEps, err := strconv.Atoi(epsStr)
		if err != nil || scrapedEps <= 0 {
			log.Errorf("invalid EPS: %s - target: %s:%s", epsStr, ticker, exchange.Title)
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

	security.STM = sql.NullString{
		String: string(models.TimingTTM),
		Valid:  true,
	}

	security.Dividend = scrapeDividend(ticker, exchange.Title, security.Typology, page)

	switch security.Typology {
	case "STOCK":
		err = models.CreateStock(database.DB, &security)
		if err != nil {
			return err
		}
	case "ETF":
		var etf models.ETF

		etf.Security = security

		aumStrElem, err := page.Element("span[title='Net Assets'] ~ span")
		if err != nil {
			log.Errorf("AUM not found in page - target: %s:%s", ticker, exchange.Title)
			etf.AUM = sql.NullInt64{
				Valid: false,
			}
		}

		aumStr := aumStrElem.MustText()
		scrapedAum, err := helpers.ParseNumberString(aumStr)
		if err != nil || scrapedAum <= 0 {
			log.Errorf("invalid AUM: %s - target: %s:%s", aumStr, ticker, exchange.Title)
			etf.AUM = sql.NullInt64{
				Valid: false,
			}
		} else {
			etf.AUM = sql.NullInt64{
				Int64: scrapedAum,
				Valid: true,
			}
		}

		erStrElem, err := page.Element("span[title='Expense Ratio (net)'] ~ span")
		if err != nil {
			log.Errorf("expense ratio not found in page - target: %s:%s", ticker, exchange.Title)
			etf.ExpenseRatio = sql.NullInt64{
				Valid: false,
			}
		}

		erStr := erStrElem.MustText()
		erStr = helpers.NormalizeFloatStrToIntStr(erStr)
		if erStr == "" {
			log.Errorf("empty expense ratio: %s - target: %s:%s", erStr, ticker, exchange.Title)
			etf.ExpenseRatio = sql.NullInt64{
				Valid: false,
			}
		} else {
			scrapedEr, err := strconv.Atoi(erStr)
			if err != nil || scrapedEr <= 0 {
				log.Errorf("invalid expense ratio: %s - target: %s:%s", erStr, ticker, exchange.Title)
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

		navStrElem, err := page.Element("span[title='NAV'] ~ span")
		if err != nil {
			log.Errorf("NAV not found in page - target: %s:%s", ticker, exchange.Title)
			etf.NAV = sql.NullInt64{
				Valid: false,
			}
		}

		navStr := navStrElem.MustText()
		navStr = helpers.NormalizeFloatStrToIntStr(navStr)
		if isAnEmptyString(navStr) {
			log.Errorf("empty NAV: %s - target: %s:%s", navStr, ticker, exchange.Title)
			etf.NAV = sql.NullInt64{
				Valid: false,
			}
		} else {
			scrapedNav, err := strconv.Atoi(navStr)
			if err != nil || scrapedNav <= 0 {
				log.Errorf("invalid NAV: %s - target: %s:%s", navStr, ticker, exchange.Title)
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

		inceptionDateStrElem, err := page.Element("section[data-testid='company-overview-card'] p[title]")
		if err != nil {
			log.Errorf("inception date not found in page - target: %s:%s", ticker, exchange.Title)
			etf.InceptionDate = sql.NullTime{
				Valid: false,
			}
		}

		inceptionDateStr := inceptionDateStrElem.MustText()
		if isAnEmptyString(inceptionDateStr) {
			log.Errorf("empty inception date: %s - target: %s:%s", inceptionDateStr, ticker, exchange.Title)
			etf.InceptionDate = sql.NullTime{
				Valid: false,
			}
		} else {
			scrapedInceptionDate, err := time.Parse("2006-01-02", inceptionDateStr)
			if err != nil {
				log.Errorf("invalid inception date: %s - target: %s:%s", inceptionDateStr, ticker, exchange.Title)
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

		relationsElementsTickers, err := page.Elements("section[data-testid='top-holdings'] a[data-testid='ticker-container']")
		if err != nil {
			log.Errorf("top holdings not found in page - target: %s:%s", ticker, exchange.Title)
		}

		relationsElementsAllocations, err := page.Elements("section[data-testid='top-holdings'] a[data-testid='ticker-container'] ~ span.data")
		if err != nil {
			log.Errorf("top holdings not found in page - target: %s:%s", ticker, exchange.Title)
		}

		for i := 0; i < len(relationsElementsTickers); i++ {
			ticker := relationsElementsTickers[i].MustText()
			ticker = strings.TrimSpace(ticker)
			allocationStr := relationsElementsAllocations[i].MustText()
			allocationStr = helpers.NormalizeFloatStrToIntStr(allocationStr)
			if isAnEmptyString(allocationStr) {
				log.Errorf("empty allocation: %s - target: %s:%s", allocationStr, ticker, exchange.Title)
				continue
			}

			scrapedAllocation, err := strconv.Atoi(allocationStr)
			if err != nil || scrapedAllocation <= 0 {
				log.Errorf("invalid allocation: %s - target: %s:%s", allocationStr, ticker, exchange.Title)
				continue
			}

			etf.RelatedSecurities = append(etf.RelatedSecurities, fmt.Sprintf("%s:%s:%d", ticker, exchange.Title, scrapedAllocation))
		}

		err = models.CreateETF(database.DB, &etf)
		if err != nil {
			return err
		}

	case "REIT":
		security.STM = sql.NullString{
			String: string(models.TimingTTM),
			Valid:  true,
		}
	default:
		return fmt.Errorf("invalid typology: %s - target: %s:%s", security.Typology, ticker, exchange.Title)
	}

	return nil
}

func scrapeDividend(ticker string, exchange string, typology string, page *rod.Page) *models.Dividend {
	//Scrape Dividend Info if any
	var dividend models.Dividend

	var yieldStr string
	if typology == "ETF" {
		yieldStrElem, err := page.Element("span[title='Yield'] ~ span")
		if err != nil {
			log.Errorf("yield not found in page - target: %s:%s", ticker, exchange)
			return nil
		} else {
			yieldStr = yieldStrElem.MustText()
		}
	} else {
		yieldStrElem, err := page.Element("span[title='Forward Dividend & Yield'] ~ span")
		if err != nil {
			log.Errorf("forward dividend & yield not found in page - target: %s:%s", ticker, exchange)
			return nil
		} else {
			yieldStr = yieldStrElem.MustText()
		}
	}

	yieldStr = helpers.NormalizeFloatStrToIntStr(yieldStr)
	if isAnEmptyString(yieldStr) {
		log.Errorf("empty yield: %s - target: %s:%s", yieldStr, ticker, exchange)
		return nil
	}

	scrapedYield, err := strconv.Atoi(yieldStr)
	if err != nil {
		log.Errorf("invalid yield: %s - target: %s:%s", yieldStr, ticker, exchange)
		return nil
	}

	if scrapedYield <= 0 {
		log.Errorf("invalid negative yield: %d - target: %s:%s", scrapedYield, ticker, exchange)
		return nil
	}

	dividend.Yield = scrapedYield

	return &dividend
}
