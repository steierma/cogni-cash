#  Web Dashboard

The administrative and analytical web interface for , built with React 19, TypeScript, and Tailwind CSS v4.

## 🎨 Design System

- **Primary Color:** Indigo (`#6366f1`)
- **Visuals:** Recharts 3 for interactive charts.
- **Icons:** Lucide React.
- **Theming:** Full light/dark mode support.

## 🚀 Key Features

- **Dynamic Analytics:** Monthly cash flow, category breakdowns, and merchant trends.
- **Transaction Inbox:** "Unreviewed Only" view for managing newly synchronized bank data.
- **Document Management:** Drag-and-drop support for bank statements, invoices, and payslips.
- **AI Categorization:** One-click batch categorization using local LLMs.
- **Reconciliation Wizard:** Smart matching of internal transfers between accounts.
- **Bank Integration:** Interactive OAuth flow for linking live bank feeds.
- **Multilingual:** Support for English, German, Spanish, and French.

## 🛠️ Development

### Prerequisites
- Node.js 22+
- Running  Backend

### Running Locally
```bash
npm install
npm run dev
```
The frontend will be available at `http://localhost:5173`.

### Environment Variables
- `VITE_API_URL`: Override the default API endpoint (default: same host as frontend).

## 🌍 Internationalization (i18n)

Translations are managed in `src/i18n/locales/`. 
- **MANDATORY:** Always add new keys to all 4 supported languages.
- **English** is the source of truth.
