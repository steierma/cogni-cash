---
name: i18n-guardian
description: Guarantees full-stack i18n completeness for CogniCash across EN, DE, ES, and FR. Use when adding or modifying user-visible strings in React or Flutter.
---

# i18n Guardian

Ensure that all user-visible strings are localized according to CogniCash's 4-language mandate.

## Mandate
- **Source of Truth:** `en` (English)
- **Supported Locales:** `en`, `de`, `es`, `fr`
- **Location:** `frontend/src/i18n/locales/` (Web) and `mobile/lib/l10n/` (Flutter)

## Workflows

### 1. Web (React)
When a new translation key is added in React (e.g., `t('page.key')`):
1. Run `python3 scripts/i18n_tool.py check`.
2. If keys are missing, use `python3 scripts/i18n_tool.py add <key> <value>` (value for EN).
3. Ensure the key is translated for `de`, `es`, and `fr`.
4. Run `python3 scripts/i18n_tool.py pretty` to keep files sorted.

### 2. Mobile (Flutter)
When a new UI string is added in Flutter:
1. Verify the key exists in `mobile/lib/l10n/app_en.arb`.
2. Ensure the key is also present in `app_de.arb`, `app_es.arb`, and `app_fr.arb`.
3. Manually verify that the ARB files are formatted correctly.

## Automation
- ALWAYS run `python3 scripts/support/i18n_tool.py check` before concluding an i18n-related task.
- Use the `fmtCurrency` and `fmtDate` formatters from `frontend/src/utils/formatters.ts` to respect `i18n.language`.
