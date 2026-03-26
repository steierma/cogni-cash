DO $$
DECLARE
    cat_income    UUID;
    cat_housing   UUID;
    cat_misc      UUID;
    cat_tech      UUID;
    cat_groceries UUID;
    cat_utilities UUID;
    stmt_giro_id    UUID;
    stmt_savings_id UUID;
    curr_date DATE := '2016-01-01';
    end_date  DATE := '2026-03-31';

    -- Variables for randomization
    recon_amount     NUMERIC(15,2);
    salary_gross     NUMERIC(15,2);
    salary_net       NUMERIC(15,2);
    invoice_amount   NUMERIC(15,2);
    groceries_amount NUMERIC(15,2);
    utilities_amount NUMERIC(15,2);
BEGIN
    -- 1. Fetch existing core categories
    SELECT id INTO cat_income  FROM categories WHERE name = 'Einkommen';
    SELECT id INTO cat_housing FROM categories WHERE name = 'Haus und Hausrat';
    SELECT id INTO cat_misc    FROM categories WHERE name = 'Sonstige Ausgaben';

    -- 2. Upsert custom categories and capture their IDs
    INSERT INTO categories (name, color) VALUES ('Tech & Software',      '#3b82f6')
        ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name
        RETURNING id INTO cat_tech;

    INSERT INTO categories (name, color) VALUES ('Groceries & Food',     '#10b981')
        ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name
        RETURNING id INTO cat_groceries;

    INSERT INTO categories (name, color) VALUES ('Utilities & Internet', '#f97316')
        ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name
        RETURNING id INTO cat_utilities;

    -- 3. Loop month by month
    WHILE curr_date <= end_date LOOP
        -- Generate random amounts for this month
        recon_amount     := round((random() * 400  + 100)::numeric, 2);
        salary_gross     := round((random() * 800  + 6000)::numeric, 2);
        salary_net       := salary_gross - round((random() * 200 + 2000)::numeric, 2);
        invoice_amount   := round((random() * 150  + 20)::numeric, 2);
        groceries_amount := round((random() * 250  + 150)::numeric, 2); -- €150–€400
        utilities_amount := round((random() * 50   + 80)::numeric, 2);  -- €80–€130

        -- ── Bank Statements ────────────────────────────────────────────────

        -- Giro account statement
        INSERT INTO bank_statements (
            id, account_holder, iban, statement_date,
            currency, content_hash, statement_type
        ) VALUES (
            gen_random_uuid(), 'Max Mustermann', 'DE12345678901234567890',
            curr_date + interval '28 days',
            'EUR', md5(gen_random_uuid()::text), 'giro'
        ) RETURNING id INTO stmt_giro_id;

        -- Credit card statement  (statement_type = 'credit_card', not 'credit')
        INSERT INTO bank_statements (
            id, account_holder, iban, statement_date,
            currency, content_hash, statement_type
        ) VALUES (
            gen_random_uuid(), 'Max Mustermann', 'DE09876543210987654321',
            curr_date + interval '28 days',
            'EUR', md5(gen_random_uuid()::text), 'credit_card'
        ) RETURNING id INTO stmt_savings_id;

        -- ── Transactions ───────────────────────────────────────────────────

        -- 1. Income
        INSERT INTO transactions (
            bank_statement_id, booking_date, valuta_date,
            description, amount, currency, transaction_type,
            category_id, content_hash, is_reconciled
        ) VALUES (
            stmt_giro_id, curr_date + interval '1 day', curr_date + interval '1 day',
            'Salary Mustermann GmbH', salary_net, 'EUR', 'credit',
            cat_income, md5(gen_random_uuid()::text), false
        );

        -- 2. Rent
        INSERT INTO transactions (
            bank_statement_id, booking_date, valuta_date,
            description, amount, currency, transaction_type,
            category_id, content_hash, is_reconciled
        ) VALUES (
            stmt_giro_id, curr_date + interval '3 days', curr_date + interval '3 days',
            'Rent Payment', -1200.00, 'EUR', 'debit',
            cat_housing, md5(gen_random_uuid()::text), false
        );

        -- 3. Utilities
        INSERT INTO transactions (
            bank_statement_id, booking_date, valuta_date,
            description, amount, currency, transaction_type,
            category_id, content_hash, is_reconciled
        ) VALUES (
            stmt_giro_id, curr_date + interval '4 days', curr_date + interval '4 days',
            'Telekom Internet & Power', -utilities_amount, 'EUR', 'debit',
            cat_utilities, md5(gen_random_uuid()::text), false
        );

        -- 4. Groceries
        INSERT INTO transactions (
            bank_statement_id, booking_date, valuta_date,
            description, amount, currency, transaction_type,
            category_id, content_hash, is_reconciled
        ) VALUES (
            stmt_giro_id, curr_date + interval '10 days', curr_date + interval '10 days',
            'REWE Supermarket', -groceries_amount, 'EUR', 'debit',
            cat_groceries, md5(gen_random_uuid()::text), false
        );

        -- 5. Tech subscription
        INSERT INTO transactions (
            bank_statement_id, booking_date, valuta_date,
            description, amount, currency, transaction_type,
            category_id, content_hash, is_reconciled
        ) VALUES (
            stmt_giro_id, curr_date + interval '12 days', curr_date + interval '12 days',
            'Hetzner Online GmbH', -invoice_amount, 'EUR', 'debit',
            cat_tech, md5(gen_random_uuid()::text), false
        );

        -- 6. Internal transfer out (Giro → savings / credit card settlement)
        INSERT INTO transactions (
            bank_statement_id, booking_date, valuta_date,
            description, amount, currency, transaction_type,
            category_id, content_hash, is_reconciled
        ) VALUES (
            stmt_giro_id, curr_date + interval '15 days', curr_date + interval '15 days',
            'Internal Transfer to Savings', -recon_amount, 'EUR', 'debit',
            cat_misc, md5(gen_random_uuid()::text), false
        );

        -- 7. Internal transfer in (on the credit card / savings side)
        INSERT INTO transactions (
            bank_statement_id, booking_date, valuta_date,
            description, amount, currency, transaction_type,
            category_id, content_hash, is_reconciled
        ) VALUES (
            stmt_savings_id, curr_date + interval '16 days', curr_date + interval '16 days',
            'Internal Transfer from Giro', recon_amount, 'EUR', 'credit',
            cat_misc, md5(gen_random_uuid()::text), false
        );

        -- ── Invoices ───────────────────────────────────────────────────────
        -- content_hash must be unique (NOT NULL UNIQUE added by migration 002)
        -- description column added by migration 002 as well
        INSERT INTO invoices (
            raw_text, vendor, amount, currency, invoice_date,
            description, category_id, content_hash
        ) VALUES (
            'Hetzner Online GmbH Cloud Server Instance — monthly fee',
            'Hetzner Online GmbH',
            invoice_amount,
            'EUR',
            curr_date + interval '5 days',
            'Cloud server hosting — ' || to_char(curr_date, 'Mon YYYY'),
            cat_tech,
            md5(gen_random_uuid()::text)
        );

        -- ── Payslips ───────────────────────────────────────────────────────
        -- The UNIQUE (period_month_num, period_year, employee_name) constraint
        -- is satisfied by the natural month-loop — each combination is unique.
        -- ON CONFLICT DO NOTHING guards against accidental re-runs.
        INSERT INTO payslips (
            original_file_name, original_file_size,
            content_hash,
            period_month_num, period_year,
            employee_name, tax_class, tax_id,
            gross_pay, net_pay, payout_amount
        ) VALUES (
            'Entgeltnachweis_' || to_char(curr_date, 'YYYY_MM') || '.pdf',
            45000,
            md5(gen_random_uuid()::text),
            EXTRACT(MONTH FROM curr_date)::INT,
            EXTRACT(YEAR  FROM curr_date)::INT,
            'Max Mustermann',
            '3',
            '12345678901',
            salary_gross,
            salary_net,
            salary_net
        ) ON CONFLICT (period_month_num, period_year, employee_name) DO NOTHING;

        -- Advance one month
        curr_date := curr_date + interval '1 month';
    END LOOP;
END $$;

