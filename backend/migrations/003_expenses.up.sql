-- Create expenses table
CREATE TABLE expenses (
  id SERIAL PRIMARY KEY,
  created_at TIMESTAMPTZ DEFAULT now(),
  updated_at TIMESTAMPTZ DEFAULT now(),
  deleted_at TIMESTAMPTZ,
  title TEXT NOT NULL,
  amount NUMERIC(14,2) NOT NULL CHECK (amount > 0),
  category TEXT,
  date DATE NOT NULL,
  type TEXT NOT NULL CHECK (type IN ('income', 'expense')),
  payment_method TEXT,
  notes TEXT,
  user_id INTEGER NOT NULL
);

-- Create indexes for better performance
CREATE INDEX idx_expenses_user_id ON expenses(user_id);
CREATE INDEX idx_expenses_user_date_type ON expenses(user_id, date DESC, type);
CREATE INDEX idx_expenses_user_category ON expenses(user_id, category);
CREATE INDEX idx_expenses_date_range ON expenses(date);
CREATE INDEX idx_expenses_deleted_at ON expenses(deleted_at);

-- Add comments for documentation
COMMENT ON TABLE expenses IS 'User expenses and income records';
COMMENT ON COLUMN expenses.type IS 'Type of transaction: income or expense';
COMMENT ON COLUMN expenses.amount IS 'Amount in the base currency (positive values)';
COMMENT ON COLUMN expenses.date IS 'Date of the expense/income';
COMMENT ON COLUMN expenses.category IS 'Category for grouping transactions';
