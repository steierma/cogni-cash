import { useEffect } from 'react';
import { BrowserRouter, Navigate, Outlet, Route, Routes } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { fetchSettings } from './api/client';
import i18n from './i18n';
import Layout from './components/Layout';
import DashboardPage from './pages/DashboardPage';
import TransactionsPage from './pages/TransactionsPage';
import InvoicesPage from './pages/InvoicesPage';
import CategoriesPage from './pages/CategoriesPage';
import SettingsPage from './pages/SettingsPage';
import ReconcilePage from './pages/ReconcilePage';
import BankStatementsPage from './pages/BankStatementsPage';
import PayslipsPage from './pages/PayslipsPage';
import LoginPage from './pages/LoginPage';
import AnalyticsPage from './pages/AnalyticsPage';
import UsersPage from './pages/UsersPage';

const isAuthenticated = () => {
    return !!localStorage.getItem('auth_token');
};

function ProtectedRoutes() {
    // Fetch settings to apply the language preference globally
    const { data: settings } = useQuery({
        queryKey: ['settings'],
        queryFn: fetchSettings,
    });

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

                {/* All routes inside here will have the Sidebar and Theme applied */}
                <Route element={<ProtectedRoutes />}>
                    <Route path="/" element={<DashboardPage />} />
                    <Route path="/analytics" element={<AnalyticsPage />} />
                    <Route path="/transactions" element={<TransactionsPage />} />
                    <Route path="/invoices" element={<InvoicesPage />} />
                    <Route path="/payslips" element={<PayslipsPage />} />
                    <Route path="/categories" element={<CategoriesPage />} />
                    <Route path="/reconcile" element={<ReconcilePage />} />
                    <Route path="/bank-statements" element={<BankStatementsPage />} />
                    <Route path="/users" element={<UsersPage />} />
                    <Route path="/settings" element={<SettingsPage />} />
                </Route>
            </Routes>
        </BrowserRouter>
    );
}