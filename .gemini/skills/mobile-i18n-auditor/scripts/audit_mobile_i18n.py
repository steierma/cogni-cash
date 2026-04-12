import os
import re
import sys
import argparse

# Regex to find L10n.t(context, 'key') or similar
L10N_KEY_REGEX = re.compile(r"L10n\.t\(context,\s*['\"]([^'\"]+)['\"]")
LANGUAGES = ['en', 'de', 'es', 'fr']

def find_l10n_keys_in_code(lib_dir):
    keys_found = set()
    for root, _, files in os.walk(lib_dir):
        for file in files:
            if file.endswith('.dart') and file != 'l10n.dart':
                file_path = os.path.join(root, file)
                try:
                    with open(file_path, 'r', encoding='utf-8') as f:
                        content = f.read()
                        matches = L10N_KEY_REGEX.findall(content)
                        for key in matches:
                            keys_found.add(key)
                except Exception as e:
                    print(f"Error reading {file_path}: {e}")
    return keys_found

def parse_l10n_file(l10n_file_path):
    translations = {lang: set() for lang in LANGUAGES}
    if not os.path.exists(l10n_file_path):
        print(f"Error: {l10n_file_path} not found.")
        sys.exit(1)

    with open(l10n_file_path, 'r', encoding='utf-8') as f:
        content = f.read()

    # Find the localizedValues map
    for lang in LANGUAGES:
        # Match from "'lang': {" to the closing "},"
        pattern = re.compile(rf"'{lang}':\s*\{{(.*?)\}},", re.DOTALL)
        match = pattern.search(content)
        if match:
            lang_content = match.group(1)
            # Find all keys in this language block: 'key': 'value'
            key_pattern = re.compile(r"['\"]([^'\"]+)['\"]:\s*['\"]")
            keys = key_pattern.findall(lang_content)
            for key in keys:
                translations[lang].add(key)
        else:
            print(f"Warning: Language block for '{lang}' not found in {l10n_file_path}.")

    return translations

def audit(lib_dir, l10n_file):
    print(f"--- Auditing Mobile i18n for: {lib_dir} ---")
    print(f"L10n File: {l10n_file}")
    
    code_keys = find_l10n_keys_in_code(lib_dir)
    print(f"Found {len(code_keys)} unique L10n keys used in code.")

    translations = parse_l10n_file(l10n_file)
    
    en_keys = translations['en']
    
    # 1. Keys in code but missing in EN translation
    missing_in_en = code_keys - en_keys
    if missing_in_en:
        print("\n❌ MISSING IN EN (but used in code):")
        for key in sorted(missing_in_en):
            print(f"  - {key}")
    else:
        print("\n✅ All keys used in code are present in the EN translation map.")

    # 2. Keys missing in other languages (present in EN but not in lang)
    failed = False
    if missing_in_en:
        failed = True

    for lang in LANGUAGES[1:]:
        missing_in_lang = en_keys - translations[lang]
        if missing_in_lang:
            print(f"\n❌ MISSING IN {lang.upper()} (present in EN):")
            for key in sorted(missing_in_lang):
                print(f"  - {key}")
            failed = True
        else:
            print(f"\n✅ {lang.upper()} translation map is complete relative to EN.")

    # 3. Keys in code but missing in any language (warnings)
    for lang in LANGUAGES:
        missing_used_in_lang = code_keys - translations[lang]
        if missing_used_in_lang and lang != 'en':
             print(f"\n⚠️  USED IN CODE BUT MISSING IN {lang.upper()}:")
             for key in sorted(missing_used_in_lang):
                 if key not in missing_in_en and key not in (en_keys - translations[lang]): 
                    print(f"  - {key}")

    if failed:
        print("\nAudit failed: Inconsistencies found.")
        return False
    
    print("\n✅ Audit passed: i18n is complete and aligned.")
    return True

if __name__ == '__main__':
    parser = argparse.ArgumentParser(description='Audit mobile i18n completeness.')
    parser.add_argument('--lib', default='mobile/lib', help='Path to lib directory')
    parser.add_argument('--l10n', default='mobile/lib/core/utils/l10n.dart', help='Path to l10n.dart file')
    
    args = parser.parse_args()
    
    if audit(args.lib, args.l10n):
        sys.exit(0)
    else:
        sys.exit(1)
