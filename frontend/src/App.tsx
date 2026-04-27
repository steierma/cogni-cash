import { useEffect } from 'react';
import { BrowserRouter, Navigate, Outlet, Route, Routes } from 'react-router-dom';
import i18n from './i18n';
import Layout from './components/Layout';
import DashboardPage from './pages/DashboardPage';
import TransactionsPage from './pages/TransactionsPage';
import InvoicesPage from './pages/InvoicesPage';
import CategoriesPage from './pages/CategoriesPage';
import SettingsPage from './pages/SettingsPage';
import ReconcilePage from './pages/ReconcilePage';
import BankStatementsPage from './pages/BankStatementsPage';
import BankConnectionsPage from './pages/BankConnectionsPage';
import PayslipsPage from './pages/PayslipsPage';
import ForecastingPage from './pages/ForecastingPage';
import LoginPage from './pages/LoginPage';
import AnalyticsPage from './pages/AnalyticsPage';
import UsersPage from './pages/UsersPage';
import SharingDashboard from './pages/SharingDashboard';
import DocumentVaultPage from './pages/DocumentVaultPage';
import TaxYearViewPage from './pages/TaxYearViewPage';
import ForgotPasswordPage from './pages/ForgotPasswordPage';
import ResetPasswordPage from './pages/ResetPasswordPage';
import SubscriptionsPage from './pages/SubscriptionsPage';
import SubscriptionDetailPage from './pages/SubscriptionDetailPage';

import { useEffectiveSettings } from './hooks/useEffectiveSettings';

const isAuthenticated = () => {
    return document.cookie.split(';').some((item) => item.trim().startsWith('cogni_cash_logged_in=true'));
};

function ProtectedRoutes() {
    // Fetch settings to apply the language preference globally
    const { data: settings } = useEffectiveSettings();

    useEffect(() => {
        if (settings?.ui_language) {
            i18n.changeLanguage(settings.ui_language);
        }
    }, [settings?.ui_language]);

    if (!isAuthenticated()) {
        return <Navigate to="/login" replace />;
    }

    return (
        <Layout>
            <Outlet />
        </Layout>
    );
}

export default function App() {
    return (
        <BrowserRouter>
            <Routes>
                <Route path="/login" element={<LoginPage />} />
                <Route path="/forgot-password" element={<ForgotPasswordPage />} />
                <Route path="/reset-password" element={<ResetPasswordPage />} />

                {/* All routes inside here will have the Sidebar and Theme applied */}
                <Route element={<ProtectedRoutes />}>
                    <Route path="/" element={<DashboardPage />} />
                    <Route path="/analytics" element={<AnalyticsPage />} />
                    <Route path="/forecasting" element={<ForecastingPage />} />
                    <Route path="/transactions" element={<TransactionsPage />} />
                    <Route path="/subscriptions" element={<SubscriptionsPage />} />
                    <Route path="/subscriptions/:id" element={<SubscriptionDetailPage />} />
                    <Route path="/invoices" element={<InvoicesPage />} />
                    <Route path="/payslips" element={<PayslipsPage />} />
                    <Route path="/documents" element={<DocumentVaultPage />} />
                    <Route path="/documents/tax/:year" element={<TaxYearViewPage />} />
                    <Route path="/categories" element={<CategoriesPage />} />
                    <Route path="/sharing" element={<SharingDashboard />} />
                    <Route path="/reconcile" element={<ReconcilePage />} />
                    <Route path="/bank-statements" element={<BankStatementsPage />} />
                    <Route path="/bank-connections" element={<BankConnectionsPage />} />
                    <Route path="/users" element={<UsersPage />} />
                    <Route path="/settings" element={<SettingsPage />} />
                </Route>
            </Routes>
        </BrowserRouter>
    );
}