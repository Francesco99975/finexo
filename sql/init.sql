-- Create a trigger function to update the updated column
CREATE OR REPLACE FUNCTION update_updated()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;


-- Create a macro to apply the trigger to a table
CREATE OR REPLACE FUNCTION apply_update_trigger(table_name TEXT)
RETURNS VOID AS $$
BEGIN
 IF NOT EXISTS (
    SELECT 1
    FROM information_schema.triggers
    WHERE trigger_schema = 'public'
      AND trigger_name = format('trigger_update_updated_%I', table_name)
  ) THEN
    EXECUTE format('
        CREATE TRIGGER trigger_update_updated_%I
        BEFORE UPDATE ON %I
        FOR EACH ROW
        EXECUTE FUNCTION update_updated()
    ', table_name, table_name);
  END IF;
END;
$$ LANGUAGE plpgsql;


CREATE TABLE IF NOT EXISTS securities (
    ticker VARCHAR(20) NOT NULL,                 -- Ticker symbol (e.g., TD)
    cc VARCHAR(3) NOT NULL,            -- Country code (e.g., CA)
    suffix VARCHAR(10) NOT NULL,                 -- Suffix (e.g., TO)
    exchange VARCHAR(50) NOT NULL,               -- Exchange (e.g., TSX)
    typology VARCHAR(10) NOT NULL CHECK (typology IN ('Stock', 'ETF', 'REIT')), -- Enum-like validation
    fullname VARCHAR(255) NOT NULL,                  -- Full name of the security
    price INT NOT NULL,                          -- Price in cents (e.g., 7612)
    pc INT NOT NULL,                           -- Price Change in cents (e.g., 112)
    ppc VARCHAR(10) NOT NULL,                 -- Price Percentage Change (e.g., 1.23)
    yrange VARCHAR(20) NOT NULL,             -- Year range (e.g., 40-80)
    drange VARCHAR(20) NOT NULL,              -- Day range (e.g., 60-70)
    marketcap VARCHAR(50),                      -- Market cap (e.g., 133.73B)
    volume INT,                                  -- Current volume
    avgvlm INT,                              -- Average volume
    beta VARCHAR(10),                            -- Beta value (e.g., 0.82)
    pclose INT NOT NULL,                          -- Last close price in cents (e.g., 7642)
    copen INT NOT NULL,                           -- Open price in cents (e.g., 7600)
    bid INT NOT NULL,                            -- Bid price in cents (e.g., 7500)
    bidsz INT,                                -- Bid size
    ask INT NOT NULL,                            -- Ask price in cents (e.g., 7800)
    asksz INT,                                -- Ask size
    currency VARCHAR(3) NOT NULL,                -- Currency code (e.g., USD, CAD)
    created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (ticker, cc)           -- Composite primary key
);

SELECT apply_update_trigger('securities');

CREATE TABLE IF NOT EXISTS stocks (
    ticker VARCHAR(20) NOT NULL,
    cc VARCHAR(3) NOT NULL,
    eps VARCHAR(10),                                -- Earnings Per Share (e.g., 7.84)
    teps VARCHAR(10) CHECK (teps IN ('FWD', 'TTM')), -- EPS type (Forward or TTM)
    pe VARCHAR(10),                                 -- Price-to-Earnings ratio (e.g., 9.75)
    tpe VARCHAR(10) CHECK (tpe IN ('FWD', 'TTM')),   -- PE type (Forward or TTM)
    PRIMARY KEY (ticker, cc),
    FOREIGN KEY (ticker, cc) REFERENCES securities (ticker, cc) ON DELETE CASCADE
);


CREATE TABLE IF NOT EXISTS etfs (
    ticker VARCHAR(20) NOT NULL,
    cc VARCHAR(3) NOT NULL,
    holdings INT NOT NULL,                 -- Number of holdings (e.g., 100)
    aum VARCHAR(50),                             -- Assets Under Management (e.g., 500M)
    er VARCHAR(10),                   -- Expense ratio (e.g., 0.5%)
    PRIMARY KEY (ticker, cc),
    FOREIGN KEY (ticker, cc) REFERENCES securities (ticker, cc) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS etf_related_securities (
    etf_ticker VARCHAR(20) NOT NULL,
    etf_cc VARCHAR(3) NOT NULL,
    related_ticker VARCHAR(20) NOT NULL,
    related_cc VARCHAR(3) NOT NULL,
    PRIMARY KEY (etf_ticker, etf_cc, related_ticker, related_cc),
    FOREIGN KEY (etf_ticker, etf_cc) REFERENCES etfs (ticker, cc) ON DELETE CASCADE,
    FOREIGN KEY (related_ticker, related_cc) REFERENCES securities (ticker, cc) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS reits (
    ticker VARCHAR(20) NOT NULL,
    cc VARCHAR(3) NOT NULL,
    ffo INT,                                      -- Funds from Operations in cents
    tffo VARCHAR(10) CHECK (tffo IN ('FWD', 'TTM')), -- FFO type
    pffo INT,                                    -- Price/FFO ratio in cents
    tpffo VARCHAR(10) CHECK (tpffo IN ('FWD', 'TTM')), -- PFFO type
    PRIMARY KEY (ticker, cc),
    FOREIGN KEY (ticker, cc) REFERENCES securities (ticker, cc) ON DELETE CASCADE
);


CREATE TABLE IF NOT EXISTS dividends (
    ticker VARCHAR(20) NOT NULL,
    cc VARCHAR(3) NOT NULL,
    rate INT NOT NULL,                            -- Dividend rate in cents
    trate VARCHAR(10) NOT NULL CHECK (trate IN ('FWD', 'TTM')), -- Rate type
    yield VARCHAR(10) NOT NULL,                  -- Dividend yield (e.g., 5.50%)
    tyield VARCHAR(10) NOT NULL CHECK (tyield IN ('FWD', 'TTM')), -- Yield type
    ap INT,                                      -- Annual payout in cents
    tap VARCHAR(10) CHECK (tap IN ('FWD', 'TTM')), -- AP type
    pr VARCHAR(10),                              -- Payout ratio (e.g., 32.75%)
    lgr VARCHAR(10),                             -- Lustrum Growth Rate (e.g., 7.15%)
    yog INT,                                     -- Years of Growth
    lad INT,                                     -- Latest Announced Dividend in cents
    frequency VARCHAR(50),                       -- Frequency (e.g., Quarterly)
    edd TIMESTAMP,                               -- Ex-Dividend Date
    pd TIMESTAMP,                                -- Payout Date
    PRIMARY KEY (ticker, cc),
    FOREIGN KEY (ticker, cc) REFERENCES securities (ticker, cc) ON DELETE CASCADE
);


CREATE INDEX IF NOT EXISTS idx_securities_ticker_country ON securities(ticker, cc);
CREATE INDEX IF NOT EXISTS idx_securities_typology ON securities(typology);
CREATE INDEX IF NOT EXISTS idx_dividends_rate ON dividends(rate);
CREATE INDEX IF NOT EXISTS idx_etfs_aum ON etfs(aum);
CREATE INDEX IF NOT EXISTS idx_stocks_eps ON stocks(eps);
