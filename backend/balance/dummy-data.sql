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
    end_date DATE := '2026-03-01';

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

    -- 1. Fetch existing core categories
    SELECT id INTO cat_income FROM categories WHERE name = 'Einkommen' AND user_id = admin_id;
    SELECT id INTO cat_housing FROM categories WHERE name = 'Haus und Hausrat' AND user_id = admin_id;
    SELECT id INTO cat_misc FROM categories WHERE name = 'Sonstige Ausgaben' AND user_id = admin_id;

    -- 2. Create and fetch new custom categories for a realistic distribution
    INSERT INTO categories (name, color, user_id, is_variable_spending) VALUES ('Tech & Software', '#3b82f6', admin_id, true)
    ON CONFLICT (name, user_id) DO UPDATE SET name=EXCLUDED.name RETURNING id INTO cat_tech;

    INSERT INTO categories (name, color, user_id, is_variable_spending) VALUES ('Groceries & Food', '#10b981', admin_id, true)
    ON CONFLICT (name, user_id) DO UPDATE SET name=EXCLUDED.name RETURNING id INTO cat_groceries;

    INSERT INTO categories (name, color, user_id, is_variable_spending) VALUES ('Utilities & Internet', '#f97316', admin_id, false)
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
        INSERT INTO transactions (user_id, bank_statement_id, booking_date, valuta_date, description, amount, transaction_type, category_id, content_hash, is_reconciled, statement_type, reviewed, counterparty_name, skip_forecasting, is_payslip_verified)
        VALUES (admin_id, stmt_giro_id, curr_date + interval '1 day', curr_date + interval '1 day', 'Salary Mustermann GmbH', salary_net, 'credit', cat_income, md5(gen_random_uuid()::text), false, 'giro', true, 'Mustermann GmbH', false, false);

        -- 2. Rent (Haus und Hausrat)
        INSERT INTO transactions (user_id, bank_statement_id, booking_date, valuta_date, description, amount, transaction_type, category_id, content_hash, is_reconciled, statement_type, reviewed, counterparty_name, skip_forecasting, is_payslip_verified)
        VALUES (admin_id, stmt_giro_id, curr_date + interval '3 days', curr_date + interval '3 days', 'Rent Payment', -1200.00, 'debit', cat_housing, md5(gen_random_uuid()::text), false, 'giro', true, 'Hausverwaltung GmbH', false, false);

        -- 3. Utilities (Utilities & Internet)
        INSERT INTO transactions (user_id, bank_statement_id, booking_date, valuta_date, description, amount, transaction_type, category_id, content_hash, is_reconciled, statement_type, reviewed, counterparty_name, skip_forecasting, is_payslip_verified)
        VALUES (admin_id, stmt_giro_id, curr_date + interval '4 days', curr_date + interval '4 days', 'Telekom Internet & Power', -utilities_amount, 'debit', cat_utilities, md5(gen_random_uuid()::text), false, 'giro', true, 'Deutsche Telekom', false, false);

        -- 4. Groceries (Groceries & Food)
        INSERT INTO transactions (user_id, bank_statement_id, booking_date, valuta_date, description, amount, transaction_type, category_id, content_hash, is_reconciled, statement_type, reviewed, counterparty_name, skip_forecasting, is_payslip_verified)
        VALUES (admin_id, stmt_giro_id, curr_date + interval '10 days', curr_date + interval '10 days', 'REWE Supermarket', -groceries_amount, 'debit', cat_groceries, md5(gen_random_uuid()::text), false, 'giro', true, 'REWE', false, false);

        -- 5. Tech & Subscriptions (Tech & Software)
        INSERT INTO transactions (user_id, bank_statement_id, booking_date, valuta_date, description, amount, transaction_type, category_id, content_hash, is_reconciled, statement_type, reviewed, counterparty_name, skip_forecasting, is_payslip_verified)
        VALUES (admin_id, stmt_giro_id, curr_date + interval '12 days', curr_date + interval '12 days', 'Hetzner Online GmbH', -invoice_amount, 'debit', cat_tech, md5(gen_random_uuid()::text), false, 'giro', true, 'Hetzner Online GmbH', false, false);

        -- ---------------------------------------------------------
        -- OPEN RECONCILIATIONS (1:1 Transfers)
        -- ---------------------------------------------------------
        -- Pair A: Multi-day offset (Random amount)
        INSERT INTO transactions (user_id, bank_statement_id, booking_date, valuta_date, description, amount, transaction_type, category_id, content_hash, is_reconciled, statement_type, reviewed, skip_forecasting, is_payslip_verified)
        VALUES (admin_id, stmt_giro_id, curr_date + interval '15 days', curr_date + interval '15 days', 'Internal Transfer to Savings', -recon_amount, 'debit', cat_misc, md5(gen_random_uuid()::text), false, 'giro', true, false, false);

        INSERT INTO transactions (user_id, bank_statement_id, booking_date, valuta_date, description, amount, transaction_type, category_id, content_hash, is_reconciled, statement_type, reviewed, skip_forecasting, is_payslip_verified)
        VALUES (admin_id, stmt_savings_id, curr_date + interval '16 days', curr_date + interval '16 days', 'Internal Transfer from Giro', recon_amount, 'credit', cat_misc, md5(gen_random_uuid()::text), false, 'credit_card', true, false, false);

        -- Pair B: Guaranteed same-day exact match (1500.00)
        INSERT INTO transactions (user_id, bank_statement_id, booking_date, valuta_date, description, amount, transaction_type, category_id, content_hash, is_reconciled, statement_type, reviewed, skip_forecasting, is_payslip_verified)
        VALUES (admin_id, stmt_giro_id, curr_date + interval '20 days', curr_date + interval '20 days', 'Manual Transfer Extra (Out)', -1500.00, 'debit', cat_misc, md5(gen_random_uuid()::text), false, 'giro', true, false, false);

        INSERT INTO transactions (user_id, bank_statement_id, booking_date, valuta_date, description, amount, transaction_type, category_id, content_hash, is_reconciled, statement_type, reviewed, skip_forecasting, is_payslip_verified)
        VALUES (admin_id, stmt_savings_id, curr_date + interval '20 days', curr_date + interval '20 days', 'Manual Transfer Extra (In)', 1500.00, 'credit', cat_misc, md5(gen_random_uuid()::text), false, 'credit_card', true, false, false);

        -- ---------------------------------------------------------
        -- Standalone Invoices & Payslips
        -- ---------------------------------------------------------
        INSERT INTO invoices (user_id, vendor, amount, invoice_date, category_id)
        VALUES (admin_id, 'Hetzner Online GmbH', invoice_amount, curr_date + interval '5 days', cat_tech);

        INSERT INTO payslips (user_id, employer_name, original_file_name, content_hash, period_month_num, period_year, tax_class, tax_id, gross_pay, net_pay, payout_amount)
        VALUES (
            admin_id,
            'Mustermann GmbH',
            'Entgeltnachweis_' || to_char(curr_date, 'YYYY_MM') || '.pdf',
            md5(gen_random_uuid()::text),
            EXTRACT(MONTH FROM curr_date),
            EXTRACT(YEAR FROM curr_date),
            '3',
            '12345678901',
            salary_gross,
            salary_net,
            salary_net
        );

        -- Advance one month
        curr_date := curr_date + interval '1 month';
    END LOOP;

    -- ---------------------------------------------------------
    -- Bank Connections & Accounts (API Sync)
    -- ---------------------------------------------------------
    INSERT INTO bank_connections (id, user_id, institution_id, institution_name, provider, requisition_id, reference_id, status, created_at, expires_at)
    VALUES (gen_random_uuid(), admin_id, 'SANDBOX_ID', 'Sandbox Bank', 'enablebanking', 'dummy_requisition', 'dummy_ref', 'linked', NOW(), NOW() + interval '90 days');

    INSERT INTO bank_accounts (id, connection_id, provider_account_id, iban, name, currency, balance, last_synced_at, account_type, last_sync_error)
    SELECT gen_random_uuid(), id, 'dummy_acc_id', 'DE12345678901234567890', 'Main Giro', 'EUR', 5000.00, NOW(), 'giro', NULL
    FROM bank_connections WHERE institution_id = 'SANDBOX_ID' AND user_id = admin_id;

    -- ---------------------------------------------------------
    -- Planned Transactions (Manual Forecasts)
    -- ---------------------------------------------------------
    INSERT INTO planned_transactions (id, user_id, amount, date, description, category_id, status)
    VALUES (gen_random_uuid(), admin_id, -800.00, (CURRENT_DATE + interval '1 month')::DATE, 'Summer Vacation', cat_misc, 'pending');

    INSERT INTO planned_transactions (id, user_id, amount, date, description, category_id, status)
    VALUES (gen_random_uuid(), admin_id, 300.00, (CURRENT_DATE + interval '2 months')::DATE, 'Tax Refund', cat_income, 'pending');

END $$;
