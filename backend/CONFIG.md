# Backend Configuration Guide

## Environment Variables

Create a `.env` file in the backend directory with the following variables:

```env
# Database Configuration
DB_DSN=host=localhost user=postgres password=your_password dbname=expense_tracker port=5432 sslmode=disable

# JWT Configuration
JWT_SECRET=your-super-secret-jwt-key-change-this-in-production

# Default Budget (in INR)
DEFAULT_BUDGET=50000

# Server Port
PORT=8080
```

## Database Setup

1. Make sure PostgreSQL is running
2. Create the database:
   ```bash
   createdb expense_tracker
   ```
3. Update the DB_DSN with your actual PostgreSQL credentials

## Default Values

- **Database**: PostgreSQL on localhost:5432
- **Database Name**: expense_tracker
- **Backend Port**: 8080
- **Default Budget**: 50,000 INR
- **JWT Secret**: Change this in production!

## Security Notes

- Always change the JWT_SECRET in production
- Use strong passwords for database
- Consider using environment-specific configurations 