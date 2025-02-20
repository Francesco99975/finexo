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

-- Enum definitions
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'timing') THEN
        CREATE TYPE timing AS ENUM ('fwd', 'ttm');
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'frequence') THEN
        CREATE TYPE frequence AS ENUM ('unknown', 'weekly', 'biweekly', 'monthly', 'quarterly', 'semi-annual', 'annual');
    END IF;
END $$;


CREATE TABLE IF NOT EXISTS exchanges (
    title VARCHAR(50) PRIMARY KEY,
    prefix VARCHAR(20),
    suffix VARCHAR(20),
    cc VARCHAR(20) NOT NULL,
    opentime TIME,
    closetime TIME
);

-- Indexes for exchanges
CREATE INDEX IF NOT EXISTS idx_exchanges_cc ON exchanges (cc);
CREATE INDEX IF NOT EXISTS idx_exchanges_prefix_suffix ON exchanges (cc, prefix, suffix);


CREATE TABLE IF NOT EXISTS securities (
    ticker VARCHAR(20) NOT NULL,
    exchange VARCHAR(50) NOT NULL,
    typology VARCHAR(10) NOT NULL CHECK (typology IN ('STOCK', 'ETF', 'REIT')),
    currency VARCHAR(10) NOT NULL,
    fullname VARCHAR(255) NOT NULL,
    sector VARCHAR,
    industry VARCHAR,
    subindustry VARCHAR,
    price INT NOT NULL,
    pc INT NOT NULL,
    pcp INT NOT NULL,
    yrl INT NOT NULL,
    yrh INT NOT NULL,
    drl INT NOT NULL,
    drh INT NOT NULL,
    consensus VARCHAR,
    score INT,
    coverage INT,
    cap BIGINT,
    volume BIGINT,
    avgvolume BIGINT,
    outstanding BIGINT,
    beta INT,
    pclose INT NOT NULL,
    copen INT NOT NULL,
    bid INT NOT NULL,
    bidsz INT,
    ask INT NOT NULL,
    asksz INT,
    eps INT,
    pe INT,
    stm timing,
    created TIMESTAMP NOT NULL DEFAULT NOW(),
    updated TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (ticker, exchange),
    FOREIGN KEY (exchange) REFERENCES exchanges(title)
);

SELECT apply_update_trigger('securities');


CREATE TABLE IF NOT EXISTS etfs (
    ticker VARCHAR(20) NOT NULL,
    exchange VARCHAR(50) NOT NULL,
    family VARCHAR NOT NULL,
    holdings INT NOT NULL,
    aum VARCHAR(50),
    er VARCHAR(10),
    nav INT,
    inception DATE,
    PRIMARY KEY (ticker, exchange),
    FOREIGN KEY (ticker, exchange) REFERENCES securities (ticker, exchange) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS etf_related_securities (
    etf_ticker VARCHAR(20) NOT NULL,
    etf_exchange VARCHAR(3) NOT NULL,
    related_ticker VARCHAR(20) NOT NULL,
    related_exchange VARCHAR(3) NOT NULL,
    allocation INT NOT NULL,
    PRIMARY KEY (etf_ticker, etf_exchange, related_ticker, related_exchange),
    FOREIGN KEY (etf_ticker, etf_exchange) REFERENCES etfs (ticker, exchange) ON DELETE CASCADE,
    FOREIGN KEY (related_ticker, related_exchange) REFERENCES securities (ticker, exchange) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS reits (
    ticker VARCHAR(20) NOT NULL,
    exchange VARCHAR(50) NOT NULL,
    occupation INT,
    focus VARCHAR(50),
    ffo INT,
    pffo INT,
    tm timing,
    PRIMARY KEY (ticker, exchange),
    FOREIGN KEY (ticker, exchange) REFERENCES securities (ticker, exchange) ON DELETE CASCADE
);


CREATE TABLE IF NOT EXISTS dividends (
    ticker VARCHAR(20) NOT NULL,
    exchange VARCHAR(50) NOT NULL,
    yield INT NOT NULL,                  -- Dividend yield (e.g., 5.50%)
    ap INT,                                      -- Annual payout in cents
    tm timing,
    pr INT,                              -- Payout ratio (e.g., 32.75%)
    lgr INT,                             -- Lustrum Growth Rate (e.g., 7.15%)
    yog INT,                                     -- Years of Growth
    lad INT,                                     -- Latest Announced Dividend in cents
    frequency frequence,                       -- Frequency (e.g., Quarterly)
    edd TIMESTAMP,                               -- Ex-Dividend Date
    pd TIMESTAMP,                                -- Payout Date
    PRIMARY KEY (ticker, exchange),
    FOREIGN KEY (ticker, exchange) REFERENCES securities (ticker, exchange) ON DELETE CASCADE
);


CREATE INDEX IF NOT EXISTS idx_securities_ticker_exchange ON securities(ticker, exchange);
CREATE INDEX IF NOT EXISTS idx_securities_typology ON securities(typology);
CREATE INDEX IF NOT EXISTS idx_dividends_yield ON dividends(yield);
CREATE INDEX IF NOT EXISTS idx_etfs_aum ON etfs(aum);
