-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create users table
CREATE TABLE users (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  email TEXT UNIQUE NOT NULL,
  password_hash TEXT NOT NULL,
  created_at TIMESTAMPTZ DEFAULT now(),
  updated_at TIMESTAMPTZ DEFAULT now()
);

-- Create bank_links table
CREATE TABLE bank_links (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  aa_consent_id TEXT NOT NULL,
  fi_type TEXT NOT NULL,         -- "SAVINGS", "CURRENT", etc.
  status TEXT NOT NULL,          -- "PENDING", "ACTIVE", "REVOKED"
  valid_till TIMESTAMPTZ,
  created_at TIMESTAMPTZ DEFAULT now(),
  updated_at TIMESTAMPTZ DEFAULT now()
);

-- Create transactions table
CREATE TABLE transactions (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  bank_link_id UUID REFERENCES bank_links(id) ON DELETE SET NULL,
  posted_at TIMESTAMPTZ NOT NULL,
  value_date TIMESTAMPTZ,
  amount NUMERIC(14,2) NOT NULL,
  currency TEXT DEFAULT 'INR',
  txn_type TEXT NOT NULL,                   -- "DEBIT" | "CREDIT"
  balance_after NUMERIC(14,2),
  description_raw TEXT,
  merchant_name TEXT,
  account_ref TEXT,                         -- masked account / VPA
  category TEXT,
  subcategory TEXT,
  hash_dedupe TEXT NOT NULL,                -- unique hash
  source_meta JSONB DEFAULT '{}'::jsonb,
  created_at TIMESTAMPTZ DEFAULT now(),
  updated_at TIMESTAMPTZ DEFAULT now()
);

-- Create unique index for deduplication
CREATE UNIQUE INDEX ux_txn_dedupe ON transactions(hash_dedupe);

-- Create category_overrides table
CREATE TABLE category_overrides (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  matcher TEXT NOT NULL,        -- regex or contains
  category TEXT NOT NULL,
  subcategory TEXT,
  created_at TIMESTAMPTZ DEFAULT now()
);

-- Create indexes for better performance
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_bank_links_user_id ON bank_links(user_id);
CREATE INDEX idx_bank_links_consent_id ON bank_links(aa_consent_id);
CREATE INDEX idx_bank_links_status ON bank_links(status);
CREATE INDEX idx_transactions_user_id ON transactions(user_id);
CREATE INDEX idx_transactions_bank_link_id ON transactions(bank_link_id);
CREATE INDEX idx_transactions_posted_at ON transactions(posted_at);
CREATE INDEX idx_transactions_txn_type ON transactions(txn_type);
CREATE INDEX idx_transactions_category ON transactions(category);
CREATE INDEX idx_category_overrides_user_id ON category_overrides(user_id);

-- Create composite indexes for common query patterns
CREATE INDEX idx_transactions_user_posted ON transactions(user_id, posted_at DESC);
CREATE INDEX idx_transactions_user_category ON transactions(user_id, category);
CREATE INDEX idx_transactions_user_type ON transactions(user_id, txn_type); 