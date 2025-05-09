# Finexo API

![Coverage](.github/badges/coverage.svg)

Finexo is a financial data API that provides information on stocks, ETFs, and REITs. It aggregates data from multiple sources, including Yahoo Finance, DividendHistory.org, and MarketBeat. The API allows users to query securities and obtain detailed financial information.

## Project Status

This project is currently in development. Features and data structures are subject to change as improvements and enhancements are made.

## Features

- Retrieve financial data on stocks, ETFs, and REITs.
- Filter securities by exchange, country, price range, and dividend status.
- Sort and limit results for better control over data retrieval.

## Available Endpoints

### Base URL

`/api/v1`

### Endpoints

#### **1. `/stocks`**

#### **2. `/etfs`**

#### **3. `/reits`**

Each of these endpoints supports query parameters to refine search results. The following parameters can be used:

- `exchange` (string, optional) – Filter by stock exchange.
- `country` (string, optional) – Filter by the country of the security.
- `minPrice` (int, optional) – Minimum price filter.
- `maxPrice` (int, optional) – Maximum price filter.
- `dividend` (bool, optional) – If set to `true`, only securities that pay dividends are returned.
- `order` (string, optional) – Specifies the field by which results should be ordered (e.g., `price`, `yield`).
- `asc` (string, optional) – Determines if results should be sorted in ascending (`true`) or descending (`false`) order.
- `limit` (int, optional) – Limits the number of returned results.

### Example Request

```http
GET /api/v1/stocks?exchange=NASDAQ&minPrice=50&maxPrice=500&dividend=true&order=yield&asc=false&limit=10
```

This request fetches up to 10 dividend-paying stocks from the NASDAQ exchange, priced between $50 and $500, sorted by dividend yield in descending order.

# Data Overview

The API provides detailed financial data, including:

- **Stocks**: Company name, sector, industry, stock price, market capitalization, trading volume, earnings per share (EPS), price-to-earnings (P/E) ratio, beta, and more.
- **ETFs**: Fund family, number of holdings, assets under management (AUM), expense ratio (ER), net asset value (NAV), and inception date.
- **REITs**: Property focus, funds from operations (FFO), price-to-FFO ratio, and other key REIT-specific metrics.
- **Dividends**: Dividend yield, annual payout, payout ratio, growth rate, ex-dividend date, payout date, and frequency.

This API is designed to help developers and financial analysts access structured data for financial securities efficiently.
