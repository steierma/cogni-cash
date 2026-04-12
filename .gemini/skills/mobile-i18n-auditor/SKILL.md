---
name: mobile-i18n-auditor
description: Audits and aligns Flutter i18n placeholders with the hardcoded l10n.dart map. Use when modifying or adding UI strings in the mobile app.
---

# Mobile i18n Auditor

This skill ensures that every `L10n.t(context, 'key')` call in the Flutter codebase has a corresponding entry in `mobile/lib/core/utils/l10n.dart` for all 4 supported languages (EN, DE, ES, FR).

## Mandate
- **Source of Truth:** `mobile/lib/core/utils/l10n.dart`
- **Supported Locales:** `en`, `de`, `es`, `fr`
- **Pattern:** `L10n.t(context, 'key')` or `L10n.t(context, "key")` or `L10n.t(context, 'key', fallback: '...')`

## Workflows

### 1. Audit Phase
When a new UI string is added or an existing one is modified:
1. Run the audit script: `python3 .gemini/skills/mobile-i18n-auditor/scripts/audit_mobile_i18n.py`.
2. Review the output for:
    - **Missing Keys:** Keys used in the code but not defined in `l10n.dart`.
    - **Incomplete Translations:** Keys defined in EN but missing in DE, ES, or FR.
    - **Unused Keys:** (Optional) Keys defined in `l10n.dart` but not found in the code.

### 2. Alignment Phase
1. Add missing keys to the `en` map in `mobile/lib/core/utils/l10n.dart`.
2. Translate the keys for `de`, `es`, and `fr`.
3. Re-run the audit script to verify completeness.

## Automation
- ALWAYS run the audit script before concluding any mobile UI task.
- Ensure that the `fallback` parameter in `L10n.t` (if used) matches the English translation for consistency.
