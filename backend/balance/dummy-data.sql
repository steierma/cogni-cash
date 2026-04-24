DO $$
DECLARE
    admin_id UUID;
    cat_income UUID;
    cat_housing UUID;
    cat_misc UUID;
    cat_tech UUID;
    cat_groceries UUID;
    cat_utilities UUID;
    stmt_giro_id UUID;
    stmt_savings_id UUID;
    curr_date DATE := '2016-01-01';
    end_date DATE := '2026-04-01';

    -- Variables for randomization
    recon_amount NUMERIC(15,2);
    base_salary_gross NUMERIC(15,2) := 5200.00; -- Starting base in 2016
    salary_gross NUMERIC(15,2);
    salary_net NUMERIC(15,2);
    invoice_amount NUMERIC(15,2);
    groceries_amount NUMERIC(15,2);
    utilities_amount NUMERIC(15,2);
BEGIN
    -- 0. Fetch Admin ID for tenancy
    SELECT id INTO admin_id FROM users WHERE username = 'admin' LIMIT 1;
    
    IF admin_id IS NULL THEN
        RAISE EXCEPTION 'Admin user not found. Please run migrations and ensure admin user exists.';
    END IF;

    -- ---------------------------------------------------------
    -- Seed 'test' User (Test / Test)
    -- ---------------------------------------------------------
    DECLARE
        test_id UUID;
        test_cat_income UUID;
        test_cat_rent UUID;
        test_cat_misc UUID;
        test_stmt_id UUID;
        test_curr_date DATE := '2018-01-01'; -- 8 years of data
    BEGIN
        INSERT INTO users (username, email, password_hash, full_name, role)
        VALUES ('test', 'test@localhost', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgNoTfQjZ4mE7.v4/e0G7X0.m1t2', 'Test User', 'manager')
        ON CONFLICT (username) DO UPDATE SET full_name = EXCLUDED.full_name
        RETURNING id INTO test_id;

        INSERT INTO categories (name, color, user_id, is_variable_spending, forecast_strategy) VALUES ('Salary', '#4caf50', test_id, false, '3y') ON CONFLICT (name, user_id) DO UPDATE SET name=EXCLUDED.name RETURNING id INTO test_cat_income;
        INSERT INTO categories (name, color, user_id, is_variable_spending, forecast_strategy) VALUES ('Monthly Rent', '#f44336', test_id, false, '3y') ON CONFLICT (name, user_id) DO UPDATE SET name=EXCLUDED.name RETURNING id INTO test_cat_rent;
        INSERT INTO categories (name, color, user_id, is_variable_spending, forecast_strategy) VALUES ('Lifestyle', '#9c27b0', test_id, true, '6m') ON CONFLICT (name, user_id) DO UPDATE SET name=EXCLUDED.name RETURNING id INTO test_cat_misc;

        WHILE test_curr_date <= end_date LOOP
            INSERT INTO bank_statements (id, user_id, account_holder, iban, statement_date, content_hash, statement_type)
            VALUES (gen_random_uuid(), test_id, 'Test User', 'DE112233445566778899', test_curr_date + interval '28 days', md5(gen_random_uuid()::text), 'giro')
            RETURNING id INTO test_stmt_id;

            INSERT INTO transactions (user_id, bank_statement_id, booking_date, valuta_date, description, amount, base_amount, base_currency, transaction_type, category_id, content_hash, statement_type, reviewed)
            VALUES (test_id, test_stmt_id, test_curr_date + interval '1 day', test_curr_date + interval '1 day', 'Salary Test Corp', 4200.00, 4200.00, 'EUR', 'credit', test_cat_income, md5(gen_random_uuid()::text), 'giro', true);

            INSERT INTO transactions (user_id, bank_statement_id, booking_date, valuta_date, description, amount, base_amount, base_currency, transaction_type, category_id, content_hash, statement_type, reviewed)
            VALUES (test_id, test_stmt_id, test_curr_date + interval '2 days', test_curr_date + interval '2 days', 'Rent Payment (Test)', -950.00, -950.00, 'EUR', 'debit', test_cat_rent, md5(gen_random_uuid()::text), 'giro', true);

            INSERT INTO transactions (user_id, bank_statement_id, booking_date, valuta_date, description, amount, base_amount, base_currency, transaction_type, category_id, content_hash, statement_type, reviewed)
            VALUES (test_id, test_stmt_id, test_curr_date + interval '15 days', test_curr_date + interval '15 days', 'Amazon.de Purchase', -120.00, -120.00, 'EUR', 'debit', test_cat_misc, md5(gen_random_uuid()::text), 'giro', true);

            test_curr_date := test_curr_date + interval '1 month';
        END LOOP;
    END;

    -- 1. Fetch existing core categories
    SELECT id INTO cat_income FROM categories WHERE name = 'Einkommen' AND user_id = admin_id;
    SELECT id INTO cat_housing FROM categories WHERE name = 'Haus und Hausrat' AND user_id = admin_id;
    SELECT id INTO cat_misc FROM categories WHERE name = 'Sonstige Ausgaben' AND user_id = admin_id;

    -- 2. Create and fetch new custom categories for a realistic distribution
    INSERT INTO categories (name, color, user_id, is_variable_spending, forecast_strategy) VALUES ('Tech & Software', '#3b82f6', admin_id, true, '6m')
    ON CONFLICT (name, user_id) DO UPDATE SET name=EXCLUDED.name RETURNING id INTO cat_tech;

    INSERT INTO categories (name, color, user_id, is_variable_spending, forecast_strategy) VALUES ('Groceries & Food', '#10b981', admin_id, true, '3m')
    ON CONFLICT (name, user_id) DO UPDATE SET name=EXCLUDED.name RETURNING id INTO cat_groceries;

    INSERT INTO categories (name, color, user_id, is_variable_spending, forecast_strategy) VALUES ('Utilities & Internet', '#f97316', admin_id, false, '3y')
    ON CONFLICT (name, user_id) DO UPDATE SET name=EXCLUDED.name RETURNING id INTO cat_utilities;

    -- 3. Loop month by month for 10 years
    WHILE curr_date <= end_date LOOP
        -- Apply yearly salary increase in May (1.5% - 3.0%)
        IF EXTRACT(MONTH FROM curr_date) = 5 THEN
            base_salary_gross := base_salary_gross * (1 + (random() * 0.015 + 0.015));
        END IF;

        -- Generate monthly amounts with small fluctuations around the base
        salary_gross := round((base_salary_gross + (random() * 100 - 50))::numeric, 2);
        -- Net is approx 62% of gross (typical German ratio)
        salary_net := round((salary_gross * 0.62)::numeric, 2);
        
        recon_amount := round((random() * 400 + 100)::numeric, 2);
        invoice_amount := round((random() * 150 + 20)::numeric, 2);
        groceries_amount := round((random() * 250 + 150)::numeric, 2);
        utilities_amount := round((random() * 50 + 80)::numeric, 2);

        -- Create a Giro Statement
        INSERT INTO bank_statements (id, user_id, account_holder, iban, statement_date, content_hash, statement_type)
        VALUES (gen_random_uuid(), admin_id, 'Max Mustermann', 'DE12345678901234567890', curr_date + interval '28 days', md5(gen_random_uuid()::text), 'giro')
        RETURNING id INTO stmt_giro_id;

        -- Create a Savings/Credit Statement
        INSERT INTO bank_statements (id, user_id, account_holder, iban, statement_date, content_hash, statement_type)
        VALUES (gen_random_uuid(), admin_id, 'Max Mustermann', 'DE09876543210987654321', curr_date + interval '28 days', md5(gen_random_uuid()::text), 'credit')
        RETURNING id INTO stmt_savings_id;

        -- ---------------------------------------------------------
        -- Insert Categorized Transactions
        -- ---------------------------------------------------------
        -- 1. Income (Einkommen)
        INSERT INTO transactions (user_id, bank_statement_id, booking_date, valuta_date, description, amount, base_amount, base_currency, transaction_type, category_id, content_hash, is_reconciled, statement_type, reviewed, counterparty_name, skip_forecasting, is_payslip_verified)
        VALUES (admin_id, stmt_giro_id, curr_date + interval '1 day', curr_date + interval '1 day', 'Salary Mustermann GmbH', salary_net, salary_net, 'EUR', 'credit', cat_income, md5(gen_random_uuid()::text), false, 'giro', true, 'Mustermann GmbH', false, false);

        -- 2. Rent (Haus und Hausrat)
        INSERT INTO transactions (user_id, bank_statement_id, booking_date, valuta_date, description, amount, base_amount, base_currency, transaction_type, category_id, content_hash, is_reconciled, statement_type, reviewed, counterparty_name, skip_forecasting, is_payslip_verified)
        VALUES (admin_id, stmt_giro_id, curr_date + interval '3 days', curr_date + interval '3 days', 'Rent Payment', -1200.00, -1200.00, 'EUR', 'debit', cat_housing, md5(gen_random_uuid()::text), false, 'giro', true, 'Hausverwaltung GmbH', false, false);

        -- 3. Utilities (Utilities & Internet)
        INSERT INTO transactions (user_id, bank_statement_id, booking_date, valuta_date, description, amount, base_amount, base_currency, transaction_type, category_id, content_hash, is_reconciled, statement_type, reviewed, counterparty_name, skip_forecasting, is_payslip_verified)
        VALUES (admin_id, stmt_giro_id, curr_date + interval '4 days', curr_date + interval '4 days', 'Telekom Internet & Power', -utilities_amount, -utilities_amount, 'EUR', 'debit', cat_utilities, md5(gen_random_uuid()::text), false, 'giro', true, 'Deutsche Telekom', false, false);

        -- 4. Groceries (Groceries & Food)
        INSERT INTO transactions (user_id, bank_statement_id, booking_date, valuta_date, description, amount, base_amount, base_currency, transaction_type, category_id, content_hash, is_reconciled, statement_type, reviewed, counterparty_name, skip_forecasting, is_payslip_verified)
        VALUES (admin_id, stmt_giro_id, curr_date + interval '10 days', curr_date + interval '10 days', 'REWE Supermarket', -groceries_amount, -groceries_amount, 'EUR', 'debit', cat_groceries, md5(gen_random_uuid()::text), false, 'giro', true, 'REWE', false, false);

        -- 5. Tech & Subscriptions (Tech & Software)
        INSERT INTO transactions (user_id, bank_statement_id, booking_date, valuta_date, description, amount, base_amount, base_currency, transaction_type, category_id, content_hash, is_reconciled, statement_type, reviewed, counterparty_name, skip_forecasting, is_payslip_verified)
        VALUES (admin_id, stmt_giro_id, curr_date + interval '12 days', curr_date + interval '12 days', 'Hetzner Online GmbH', -25.00, -25.00, 'EUR', 'debit', cat_tech, md5(gen_random_uuid()::text), false, 'giro', true, 'Hetzner Online GmbH', false, false);

        INSERT INTO transactions (user_id, bank_statement_id, booking_date, valuta_date, description, amount, base_amount, base_currency, transaction_type, category_id, content_hash, is_reconciled, statement_type, reviewed, counterparty_name, skip_forecasting, is_payslip_verified)
        VALUES (admin_id, stmt_giro_id, curr_date + interval '5 days', curr_date + interval '5 days', 'Netflix.com', -17.99, -17.99, 'EUR', 'debit', cat_misc, md5(gen_random_uuid()::text), false, 'giro', true, 'Netflix', false, false);

        -- ---------------------------------------------------------
        -- OPEN RECONCILIATIONS (1:1 Transfers)
        -- ---------------------------------------------------------
        -- Pair A: Multi-day offset (Random amount)
        INSERT INTO transactions (user_id, bank_statement_id, booking_date, valuta_date, description, amount, base_amount, base_currency, transaction_type, category_id, content_hash, is_reconciled, statement_type, reviewed, skip_forecasting, is_payslip_verified)
        VALUES (admin_id, stmt_giro_id, curr_date + interval '15 days', curr_date + interval '15 days', 'Internal Transfer to Savings', -recon_amount, -recon_amount, 'EUR', 'debit', cat_misc, md5(gen_random_uuid()::text), false, 'giro', true, false, false);

        INSERT INTO transactions (user_id, bank_statement_id, booking_date, valuta_date, description, amount, base_amount, base_currency, transaction_type, category_id, content_hash, is_reconciled, statement_type, reviewed, skip_forecasting, is_payslip_verified)
        VALUES (admin_id, stmt_savings_id, curr_date + interval '16 days', curr_date + interval '16 days', 'Internal Transfer from Giro', recon_amount, recon_amount, 'EUR', 'credit', cat_misc, md5(gen_random_uuid()::text), false, 'credit_card', true, false, false);

        -- Pair B: Guaranteed same-day exact match (1500.00)
        INSERT INTO transactions (user_id, bank_statement_id, booking_date, valuta_date, description, amount, base_amount, base_currency, transaction_type, category_id, content_hash, is_reconciled, statement_type, reviewed, skip_forecasting, is_payslip_verified)
        VALUES (admin_id, stmt_giro_id, curr_date + interval '20 days', curr_date + interval '20 days', 'Manual Transfer Extra (Out)', -1500.00, -1500.00, 'EUR', 'debit', cat_misc, md5(gen_random_uuid()::text), false, 'giro', true, false, false);

        INSERT INTO transactions (user_id, bank_statement_id, booking_date, valuta_date, description, amount, base_amount, base_currency, transaction_type, category_id, content_hash, is_reconciled, statement_type, reviewed, skip_forecasting, is_payslip_verified)
        VALUES (admin_id, stmt_savings_id, curr_date + interval '20 days', curr_date + interval '20 days', 'Manual Transfer Extra (In)', 1500.00, 1500.00, 'EUR', 'credit', cat_misc, md5(gen_random_uuid()::text), false, 'credit_card', true, false, false);

        -- ---------------------------------------------------------
        -- Standalone Invoices & Payslips
        -- ---------------------------------------------------------
        INSERT INTO invoices (user_id, vendor, amount, base_amount, base_currency, invoice_date, category_id)
        VALUES (admin_id, 'Hetzner Online GmbH', invoice_amount, invoice_amount, 'EUR', curr_date + interval '5 days', cat_tech);

        INSERT INTO payslips (user_id, employer_name, original_file_name, content_hash, period_month_num, period_year, tax_class, tax_id, currency, gross_pay, base_gross_pay, net_pay, base_net_pay, payout_amount, base_payout_amount)
        VALUES (
            admin_id,
            'Mustermann GmbH',
            'Entgeltnachweis_' || to_char(curr_date, 'YYYY_MM') || '.pdf',
            md5(gen_random_uuid()::text),
            EXTRACT(MONTH FROM curr_date),
            EXTRACT(YEAR FROM curr_date),
            '3',
            '12345678901',
            'EUR',
            salary_gross,
            salary_gross,
            salary_net,
            salary_net,
            salary_net,
            salary_net
        );

        -- Advance one month
        curr_date := curr_date + interval '1 month';
    END LOOP;

    -- ---------------------------------------------------------
    -- Bank Connections & Accounts (API Sync & Virtual)
    -- ---------------------------------------------------------
    INSERT INTO bank_connections (id, user_id, institution_id, institution_name, provider, requisition_id, reference_id, status, created_at, expires_at)
    VALUES (gen_random_uuid(), admin_id, 'SANDBOX_ID', 'Sandbox Bank', 'enablebanking', 'dummy_requisition', 'dummy_ref', 'linked', NOW(), NOW() + interval '90 days');

    INSERT INTO bank_accounts (id, connection_id, provider_account_id, iban, name, user_id, currency, balance, last_synced_at, account_type, last_sync_error)
    SELECT gen_random_uuid(), id, 'dummy_acc_id', 'DE12345678901234567890', 'Main Giro', admin_id, 'EUR', 5000.00, NOW(), 'giro', NULL
    FROM bank_connections WHERE institution_id = 'SANDBOX_ID' AND user_id = admin_id;

    -- Add a Virtual Account (e.g. Amazon Visa)
    INSERT INTO bank_accounts (id, connection_id, provider_account_id, iban, name, user_id, currency, balance, account_type)
    VALUES (gen_random_uuid(), NULL, 'virtual_visa_123', 'DE_VIRTUAL_VISA', 'Amazon Visa (Virtual)', admin_id, 'EUR', -450.00, 'credit_card');

    -- ---------------------------------------------------------
    -- Planned Transactions (Manual Forecasts)
    -- ---------------------------------------------------------
    INSERT INTO planned_transactions (id, user_id, amount, currency, base_amount, base_currency, date, description, category_id, status, interval_months, end_date)
    VALUES (gen_random_uuid(), admin_id, -800.00, 'EUR', -800.00, 'EUR', (CURRENT_DATE + interval '1 month')::DATE, 'Summer Vacation', cat_misc, 'pending', 0, NULL);

    INSERT INTO planned_transactions (id, user_id, amount, currency, base_amount, base_currency, date, description, category_id, status, interval_months, end_date)
    VALUES (gen_random_uuid(), admin_id, 300.00, 'EUR', 300.00, 'EUR', (CURRENT_DATE + interval '2 months')::DATE, 'Tax Refund', cat_income, 'pending', 0, NULL);

    INSERT INTO planned_transactions (id, user_id, amount, currency, base_amount, base_currency, date, description, category_id, status, interval_months, end_date)
    VALUES (gen_random_uuid(), admin_id, -50.00, 'EUR', -50.00, 'EUR', (CURRENT_DATE + interval '15 days')::DATE, 'Recurring Subscription', cat_misc, 'pending', 1, (CURRENT_DATE + interval '1 year')::DATE);

    -- ---------------------------------------------------------
    -- Subscriptions (New Feature Sample)
    -- ---------------------------------------------------------
    INSERT INTO subscriptions (user_id, merchant_name, amount, billing_cycle, billing_interval, category_id, status, last_occurrence, next_occurrence, matching_hashes, ignored_hashes, linked_mandates, linked_ibans, bank_account_id)
    SELECT admin_id, 'Hetzner Online GmbH', -25.00, 'monthly', 1, cat_tech, 'active', CURRENT_DATE - interval '12 days', CURRENT_DATE + interval '18 days', '{}', '{}', '{}', '{}', id
    FROM bank_accounts WHERE user_id = admin_id AND iban = 'DE12345678901234567890' LIMIT 1;

    INSERT INTO subscriptions (user_id, merchant_name, amount, billing_cycle, billing_interval, category_id, status, last_occurrence, next_occurrence, matching_hashes, ignored_hashes, linked_mandates, linked_ibans)
    VALUES (admin_id, 'Netflix', -17.99, 'monthly', 1, cat_misc, 'active', CURRENT_DATE - interval '5 days', CURRENT_DATE + interval '25 days', '{}', '{}', '{}', '{}');

END $$;

-- ---------------------------------------------------------
-- Shared Categories & Collaborative Data
-- ---------------------------------------------------------
DO $$
DECLARE
    admin_id UUID;
    demo_id UUID;
    shared_cat_id UUID;
    stmt_id UUID;
BEGIN
    -- 1. Ensure 'demo' user exists for sharing example
    -- password is 'password' hashed with bcrypt (cost 10)
    INSERT INTO users (username, email, password_hash, full_name, role)
    VALUES ('demo', 'demo@localhost', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgNoTfQjZ4mE7.v4/e0G7X0.m1t2', 'Demo User', 'manager')
    ON CONFLICT (username) DO UPDATE SET full_name = EXCLUDED.full_name
    RETURNING id INTO demo_id;

    SELECT id INTO admin_id FROM users WHERE username = 'admin' LIMIT 1;
    
    IF demo_id IS NULL THEN
        SELECT id INTO demo_id FROM users WHERE username = 'demo' LIMIT 1;
    END IF;

    -- If admin still doesn't exist (unexpected), skip
    IF admin_id IS NULL OR demo_id IS NULL THEN
        RETURN;
    END IF;

    -- 1. Identify the 'Groceries & Food' category owned by admin
    SELECT id INTO shared_cat_id FROM categories WHERE name = 'Groceries & Food' AND user_id = admin_id;

    -- Fallback if it doesn't exist (e.g. fresh DB without previous seeding)
    IF shared_cat_id IS NULL THEN
        INSERT INTO categories (name, color, user_id, is_variable_spending, forecast_strategy)
        VALUES ('Groceries & Food', '#10b981', admin_id, true, '3m')
        ON CONFLICT DO NOTHING RETURNING id INTO shared_cat_id;
        
        -- If still NULL, just fetch it
        IF shared_cat_id IS NULL THEN
            SELECT id INTO shared_cat_id FROM categories WHERE name = 'Groceries & Food' AND user_id = admin_id;
        END IF;
    END IF;

    -- 2. Create the sharing relationship
    INSERT INTO shared_categories (category_id, owner_user_id, shared_with_user_id, permission_level)
    VALUES (shared_cat_id, admin_id, demo_id, 'edit')
    ON CONFLICT DO NOTHING;

    -- 3. Add some transactions for 'demo' using admin's category (Live Feed)
    -- Demo paid for Pizza
    INSERT INTO transactions (user_id, bank_statement_id, booking_date, valuta_date, description, amount, base_amount, base_currency, transaction_type, category_id, content_hash, reviewed, statement_type)
    VALUES (demo_id, NULL, '2026-04-11', '2026-04-11', 'Pizza Night (Shared)', -45.50, -45.50, 'EUR', 'debit', shared_cat_id, md5('demo_tx_1'), true, 'giro')
    ON CONFLICT DO NOTHING;

    -- Demo bought drinks
    INSERT INTO transactions (user_id, bank_statement_id, booking_date, valuta_date, description, amount, base_amount, base_currency, transaction_type, category_id, content_hash, reviewed, statement_type)
    VALUES (demo_id, NULL, '2026-04-08', '2026-04-08', 'Shared Drinks & Snacks', -22.10, -22.10, 'EUR', 'debit', shared_cat_id, md5('demo_tx_2'), true, 'giro')
    ON CONFLICT DO NOTHING;

    -- 4. Add some transactions for 'admin' using the same shared category (Live Feed)
    -- This ensures 'demo' sees 'admin's' transactions when include_shared=true
    INSERT INTO transactions (user_id, bank_statement_id, booking_date, valuta_date, description, amount, base_amount, base_currency, transaction_type, category_id, content_hash, reviewed, statement_type)
    VALUES (admin_id, NULL, '2026-04-12', '2026-04-12', 'Weekly Groceries (Admin)', -89.90, -89.90, 'EUR', 'debit', shared_cat_id, md5('admin_shared_tx_1'), true, 'giro')
    ON CONFLICT DO NOTHING;

    INSERT INTO transactions (user_id, bank_statement_id, booking_date, valuta_date, description, amount, base_amount, base_currency, transaction_type, category_id, content_hash, reviewed, statement_type)
    VALUES (admin_id, NULL, '2026-04-05', '2026-04-05', 'Shared Household Items', -34.20, -34.20, 'EUR', 'debit', shared_cat_id, md5('admin_shared_tx_2'), true, 'giro')
    ON CONFLICT DO NOTHING;

    -- 5. Create a shared invoice
    -- Admin shares an invoice with Demo
    SELECT id INTO stmt_id FROM invoices WHERE user_id = admin_id LIMIT 1;
    IF stmt_id IS NOT NULL THEN
        INSERT INTO shared_invoices (invoice_id, owner_user_id, shared_with_user_id, permission_level)
        VALUES (stmt_id, admin_id, demo_id, 'view')
        ON CONFLICT DO NOTHING;
    END IF;

    -- 6. Seed Discovery Whitelisting & Caching
    -- Add a manually ignored pattern
    INSERT INTO subscription_discovery_feedback (user_id, merchant_name, status, source)
    VALUES (admin_id, 'ALDI Sued', 'DECLINED', 'USER')
    ON CONFLICT DO NOTHING;

    -- Add an AI-filtered pattern
    INSERT INTO subscription_discovery_feedback (user_id, merchant_name, status, source)
    VALUES (admin_id, 'REWE Markt GmbH', 'AI_REJECTED', 'AI')
    ON CONFLICT DO NOTHING;

    -- Add a whitelisted pattern (restored or AI-verified)
    INSERT INTO subscription_discovery_feedback (user_id, merchant_name, status, source)
    VALUES (admin_id, 'Netflix.com', 'ALLOWED', 'AI')
    ON CONFLICT DO NOTHING;

    -- 7. Create a Shared Bank Account relationship
    -- Admin shares the 'Main Giro' with Demo
    DECLARE
        shared_acc_id UUID;
    BEGIN
        SELECT id INTO shared_acc_id FROM bank_accounts WHERE user_id = admin_id AND iban = 'DE12345678901234567890' LIMIT 1;
        IF shared_acc_id IS NOT NULL THEN
            INSERT INTO shared_bank_accounts (bank_account_id, owner_user_id, shared_with_user_id, permission_level)
            VALUES (shared_acc_id, admin_id, demo_id, 'view')
            ON CONFLICT DO NOTHING;
        END IF;
    END;

END $$;
