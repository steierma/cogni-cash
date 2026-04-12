import json
import re
import os
import sys
import argparse

LANGS = ['en', 'de', 'es', 'fr']
LOCALES_DIR = 'frontend/src/i18n/locales'
SRC_DIR = 'frontend/src'

def load_json(lang):
    path = os.path.join(LOCALES_DIR, lang, 'translation.json')
    if not os.path.exists(path):
        return {}
    with open(path, 'r', encoding='utf-8') as f:
        return json.load(f)

def save_json(lang, data):
    path = os.path.join(LOCALES_DIR, lang, 'translation.json')
    os.makedirs(os.path.dirname(path), exist_ok=True)
    with open(path, 'w', encoding='utf-8') as f:
        json.dump(data, f, ensure_ascii=False, indent=2, sort_keys=True)
        f.write('\n')

def get_flat_keys(data, prefix=''):
    keys = {}
    if isinstance(data, dict):
        for k, v in data.items():
            new_prefix = f"{prefix}.{k}" if prefix else k
            if isinstance(v, dict):
                keys.update(get_flat_keys(v, new_prefix))
            else:
                keys[new_prefix] = v
    return keys

def set_nested(data, key, value):
    parts = key.split('.')
    for part in parts[:-1]:
        if part not in data or not isinstance(data[part], dict):
            data[part] = {}
        data = data[part]
    data[parts[-1]] = value

def find_keys_in_code():
    t_pattern = re.compile(r"\bt\(['\"]([^'\"$]+)['\"]")
    found_keys = set()
    for root, dirs, files in os.walk(SRC_DIR):
        if 'node_modules' in dirs: dirs.remove('node_modules')
        for file in files:
            if file.endswith(('.tsx', '.ts')):
                with open(os.path.join(root, file), 'r', encoding='utf-8') as f:
                    content = f.read()
                    for m in t_pattern.findall(content):
                        if len(m) > 1 and '${' not in m:
                            found_keys.add(m)
    return found_keys

def cmd_check():
    code_keys = find_keys_in_code()
    all_ok = True
    for lang in LANGS:
        data = load_json(lang)
        flat = get_flat_keys(data)
        missing = [k for k in code_keys if k not in flat]
        if missing:
            print(f"\n❌ Missing in {lang.upper()}:")
            for m in sorted(missing): print(f"  - {m}")
            all_ok = False
        else:
            print(f"✅ {lang.upper()} is complete.")
    return all_ok

def cmd_add(key, translations):
    for lang in LANGS:
        data = load_json(lang)
        val = translations.get(lang, translations.get('en', f"MISSING: {key}"))
        set_nested(data, key, val)
        save_json(lang, data)
    print(f"✨ Added '{key}' to all languages.")

def cmd_sync():
    en_data = load_json('en')
    en_flat = get_flat_keys(en_data)
    for lang in [l for l in LANGS if l != 'en']:
        lang_data = load_json(lang)
        lang_flat = get_flat_keys(lang_data)
        # Add missing
        for k, v in en_flat.items():
            if k not in lang_flat:
                set_nested(lang_data, k, f"[MISSING] {v}")
        # Remove extras
        new_data = {}
        for k, v in en_flat.items():
            curr_val = get_flat_keys(lang_data).get(k)
            set_nested(new_data, k, curr_val if curr_val else f"[MISSING] {v}")
        save_json(lang, new_data)
    print("✨ Synchronized all languages with EN structure.")

def main():
    parser = argparse.ArgumentParser(description="CogniCash i18n Management Tool")
    subparsers = parser.add_subparsers(dest="command")

    subparsers.add_parser("check", help="Check for missing keys")
    
    add_parser = subparsers.add_parser("add", help="Add a new key")
    add_parser.add_argument("--key", required=True)
    add_parser.add_argument("--en", required=True)
    add_parser.add_argument("--de")
    add_parser.add_argument("--es")
    add_parser.add_argument("--fr")

    subparsers.add_parser("sync", help="Sync structure from EN to all")
    subparsers.add_parser("pretty", help="Format and sort JSON files")

    args = parser.parse_args()

    if args.command == "check":
        if not cmd_check(): sys.exit(1)
    elif args.command == "add":
        tx = {'en': args.en, 'de': args.de, 'es': args.es, 'fr': args.fr}
        cmd_add(args.key, {k: v for k, v in tx.items() if v})
    elif args.command == "sync":
        cmd_sync()
    elif args.command == "pretty":
        for lang in LANGS: save_json(lang, load_json(lang))
        print("✨ All files formatted and sorted.")
    else:
        parser.print_help()

if __name__ == "__main__":
    main()
