-- Create bank_accounts table
CREATE TABLE bank_accounts (
  id SERIAL PRIMARY KEY,
  user_id INTEGER NOT NULL,
  bank_id TEXT NOT NULL,
  account_number TEXT NOT NULL,
  account_holder_name TEXT NOT NULL,
  mobile_number TEXT NOT NULL,
  status TEXT DEFAULT 'ACTIVE',
  account_type TEXT,
  ifsc_code TEXT,
  branch_name TEXT,
  verified_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ DEFAULT now(),
  updated_at TIMESTAMPTZ DEFAULT now()
);

-- Create indexes for better performance
CREATE INDEX idx_bank_accounts_user_id ON bank_accounts(user_id);
CREATE INDEX idx_bank_accounts_bank_id ON bank_accounts(bank_id);
CREATE INDEX idx_bank_accounts_status ON bank_accounts(status);

-- Create unique constraint to prevent duplicate accounts
CREATE UNIQUE INDEX ux_bank_accounts_user_bank_account ON bank_accounts(user_id, bank_id, account_number); 