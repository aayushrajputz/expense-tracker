# 💰 Smart Expense Tracker

A full-stack expense tracking application with AI-powered insights, built with Next.js 14 and Go.

## 🚀 Features

- **📊 Dashboard**: Real-time expense overview with charts and insights
- **🏦 Bank Integration**: Connect bank accounts and fetch transactions
- **🤖 AI Insights**: Get intelligent spending analysis and recommendations
- **📱 Responsive Design**: Works perfectly on desktop and mobile
- **🌍 Multi-language**: Support for English and Spanish
- **🔐 Secure Authentication**: JWT-based authentication with email verification
- **📈 Analytics**: Detailed spending analytics and reports

## 🛠️ Tech Stack

### Frontend
- **Next.js 14** (App Router)
- **React 18** with TypeScript
- **Tailwind CSS** for styling
- **Chart.js** for data visualization
- **Axios** for API calls

### Backend
- **Go** with Gin framework
- **PostgreSQL** database
- **GORM** ORM
- **JWT** authentication
- **Nodemailer** for email verification

## 📁 Project Structure

```
expense-tracker/
├── frontend/                 # Next.js frontend application
│   ├── app/                 # App Router pages
│   ├── components/          # Reusable React components
│   ├── lib/                 # Utilities and API client
│   └── package.json
├── backend/                 # Go backend application
│   ├── cmd/                 # Application entry points
│   ├── controllers/         # HTTP handlers
│   ├── models/              # Database models
│   ├── services/            # Business logic
│   ├── routes/              # API routes
│   └── go.mod
└── README.md
```

## 🚀 Quick Start

### Prerequisites
- Node.js 18+ and npm
- Go 1.22+
- PostgreSQL 14+

### Frontend Setup
```bash
cd frontend
npm install
npm run dev
```

### Backend Setup
```bash
cd backend
go mod download
go run cmd/server/main.go
```

### Environment Variables
Copy `backend/.env.example` to `backend/.env` and configure:
- Database connection
- JWT secret
- SMTP settings for email verification

## 📖 Documentation

- [Frontend README](frontend/README.md)
- [Backend README](backend/README.md)
- [API Documentation](backend/docs/README.md)

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Submit a pull request

## 📄 License

MIT License - see LICENSE file for details.

## 🆘 Support

For support, please open an issue on GitHub or contact the development team. 