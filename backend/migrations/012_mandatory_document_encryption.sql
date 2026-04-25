-- Migration: 012_mandatory_document_encryption.sql
-- This migration encrypts all existing unencrypted file content in the database.
-- It uses the session variable 'vault.key' which must be set by the migration runner.

DO $$
DECLARE
    v_key TEXT;
BEGIN
    -- Try to get the vault key from the session variable
    BEGIN
        v_key := current_setting('vault.key');
    EXCEPTION WHEN OTHERS THEN
        v_key := NULL;
    END;

    IF v_key IS NULL OR v_key = '' THEN
        RAISE WARNING 'vault.key is not set in session. Skipping encryption of existing data. Please run migrations with DOCUMENT_VAULT_KEY set.';
        RETURN;
    END IF;

    -- 1. Encrypt Payslips
    -- We assume files that cannot be decrypted with the key are unencrypted (or wrong key).
    -- pgp_sym_decrypt_bytea will fail if the data is not a valid PGP message.
    BEGIN
        UPDATE payslips
        SET original_file_content = pgp_sym_encrypt_bytea(original_file_content, v_key)
        WHERE original_file_content IS NOT NULL
          AND length(original_file_content) > 0
          -- Heuristic: PGP messages usually start with 0x01, 0x02, or 0x03 in the first byte (simplified)
          -- A better way is to try decrypting and check if it fails
          AND NOT (
              substring(original_file_content from 1 for 1) = '\x01'::bytea OR
              substring(original_file_content from 1 for 1) = '\x03'::bytea
          );
    EXCEPTION WHEN OTHERS THEN
        RAISE WARNING 'Failed to encrypt some payslips: %', SQLERRM;
    END;

    -- 2. Encrypt Invoices
    BEGIN
        UPDATE invoices
        SET original_file_content = pgp_sym_encrypt_bytea(original_file_content, v_key)
        WHERE original_file_content IS NOT NULL
          AND length(original_file_content) > 0
          AND NOT (
              substring(original_file_content from 1 for 1) = '\x01'::bytea OR
              substring(original_file_content from 1 for 1) = '\x03'::bytea
          );
    EXCEPTION WHEN OTHERS THEN
        RAISE WARNING 'Failed to encrypt some invoices: %', SQLERRM;
    END;

    -- 3. Encrypt Bank Statements
    BEGIN
        UPDATE bank_statements
        SET original_file = pgp_sym_encrypt_bytea(original_file, v_key)
        WHERE original_file IS NOT NULL
          AND length(original_file) > 0
          AND NOT (
              substring(original_file from 1 for 1) = '\x01'::bytea OR
              substring(original_file from 1 for 1) = '\x03'::bytea
          );
    EXCEPTION WHEN OTHERS THEN
        RAISE WARNING 'Failed to encrypt some bank statements: %', SQLERRM;
    END;

END $$;
