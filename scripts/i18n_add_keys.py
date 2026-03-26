#!/usr/bin/env python3
"""
i18n_add_keys.py — Generic helper for bulk-adding translation keys to the
frontend locale catalogues.

Usage examples
--------------
# Add a single flat key to every locale under the "invoices" namespace:
python3 scripts/i18n_add_keys.py \
    --namespace invoices \
    --key previewFailed \
    --translations en="Could not load preview." de="Vorschau fehlgeschlagen." \
                   es="No se pudo cargar la vista previa." fr="Impossible de charger l'aperçu."

# Apply a pre-defined "patch" JSON file (recommended for large batches):
#   patch.json looks like:
#   {
#     "en": { "invoices": { "newKey": "English text" } },
#     "de": { "invoices": { "newKey": "Deutscher Text" } }
#   }
python3 scripts/i18n_add_keys.py --patch patch.json

Configuration
-------------
LOCALES_DIR  : path to the locales root (contains en/, de/, … sub-dirs).
               Default: frontend/src/i18n/locales  (relative to repo root)
LOCALE_LANGS : comma-separated list of language codes to update.
               Default: en,de,es,fr
TRANSLATION_FILE : filename inside each lang directory.
               Default: translation.json
"""

from __future__ import annotations

import argparse
import json
import sys
from pathlib import Path


# ── Defaults (override via env-vars or CLI flags) ────────────────────────────

REPO_ROOT = Path(__file__).resolve().parent.parent
DEFAULT_LOCALES_DIR = REPO_ROOT / "frontend" / "src" / "i18n" / "locales"
DEFAULT_LANGS = ["en", "de", "es", "fr"]
DEFAULT_TRANSLATION_FILE = "translation.json"


# ── Core helpers ─────────────────────────────────────────────────────────────

def load_catalogue(path: Path) -> dict:
    with path.open("r", encoding="utf-8") as fh:
        return json.load(fh)


def save_catalogue(path: Path, data: dict) -> None:
    with path.open("w", encoding="utf-8") as fh:
        json.dump(data, fh, ensure_ascii=False, indent=2)
        fh.write("\n")  # POSIX newline at EOF


def set_nested(obj: dict, namespace: str, key: str, value: str) -> None:
    """Set obj[namespace][key] = value, creating the namespace dict if absent."""
    if namespace not in obj:
        obj[namespace] = {}
    obj[namespace][key] = value


def deep_merge(base: dict, patch: dict) -> dict:
    """Recursively merge *patch* into *base* (patch wins on conflicts)."""
    result = dict(base)
    for k, v in patch.items():
        if k in result and isinstance(result[k], dict) and isinstance(v, dict):
            result[k] = deep_merge(result[k], v)
        else:
            result[k] = v
    return result


# ── CLI ───────────────────────────────────────────────────────────────────────

def parse_args() -> argparse.Namespace:
    p = argparse.ArgumentParser(
        description="Add/update i18n keys across all locale catalogues.",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog=__doc__,
    )

    p.add_argument(
        "--locales-dir",
        type=Path,
        default=DEFAULT_LOCALES_DIR,
        help=f"Root directory that contains the per-language sub-folders. "
             f"Default: {DEFAULT_LOCALES_DIR}",
    )
    p.add_argument(
        "--langs",
        nargs="+",
        default=DEFAULT_LANGS,
        metavar="LANG",
        help=f"Language codes to update. Default: {' '.join(DEFAULT_LANGS)}",
    )
    p.add_argument(
        "--translation-file",
        default=DEFAULT_TRANSLATION_FILE,
        metavar="FILENAME",
        help=f"JSON file name inside each lang dir. Default: {DEFAULT_TRANSLATION_FILE}",
    )

    # ── Mode A: single key ───────────────────────────────────────────────────
    single = p.add_argument_group("Single-key mode")
    single.add_argument(
        "--namespace", "-n",
        help="Top-level key in the catalogue (e.g. 'invoices', 'common').",
    )
    single.add_argument(
        "--key", "-k",
        help="Dot-separated sub-key to set (e.g. 'previewFailed').",
    )
    single.add_argument(
        "--translations", "-t",
        nargs="+",
        metavar="LANG=VALUE",
        help="Per-language values as LANG=VALUE pairs (e.g. en='Hello' de='Hallo').",
    )

    # ── Mode B: patch file ───────────────────────────────────────────────────
    patch_grp = p.add_argument_group("Patch-file mode")
    patch_grp.add_argument(
        "--patch",
        type=Path,
        metavar="PATCH_JSON",
        help="Path to a JSON file shaped as {lang: {namespace: {key: value}}}.",
    )

    p.add_argument(
        "--dry-run",
        action="store_true",
        help="Print the result without writing files.",
    )

    return p.parse_args()


# ── Main ──────────────────────────────────────────────────────────────────────

def main() -> int:
    args = parse_args()

    locales_dir: Path = args.locales_dir
    if not locales_dir.is_dir():
        print(f"ERROR: locales directory not found: {locales_dir}", file=sys.stderr)
        return 1

    # Build the patch dict regardless of mode
    patch: dict[str, dict] = {}

    if args.patch:
        # Mode B: load from file
        with args.patch.open("r", encoding="utf-8") as fh:
            patch = json.load(fh)
    elif args.namespace and args.key and args.translations:
        # Mode A: single key from CLI args
        translations: dict[str, str] = {}
        for item in args.translations:
            lang, _, value = item.partition("=")
            translations[lang.strip()] = value.strip()

        for lang, value in translations.items():
            patch.setdefault(lang, {})
            set_nested(patch[lang], args.namespace, args.key, value)
    else:
        print(
            "ERROR: provide either --patch or all of --namespace / --key / --translations.",
            file=sys.stderr,
        )
        return 1

    # Apply patch to each requested language catalogue
    updated: list[str] = []
    skipped: list[str] = []

    for lang in args.langs:
        catalogue_path = locales_dir / lang / args.translation_file
        if not catalogue_path.exists():
            print(f"  SKIP  [{lang}] — file not found: {catalogue_path}", file=sys.stderr)
            skipped.append(lang)
            continue

        catalogue = load_catalogue(catalogue_path)

        if lang in patch:
            catalogue = deep_merge(catalogue, patch[lang])
        else:
            print(f"  SKIP  [{lang}] — no patch entry for this language.")
            skipped.append(lang)
            continue

        if args.dry_run:
            print(f"\n--- DRY-RUN [{lang}] ---")
            print(json.dumps(catalogue, ensure_ascii=False, indent=2))
        else:
            save_catalogue(catalogue_path, catalogue)
            print(f"  OK    [{lang}] {catalogue_path}")
            updated.append(lang)

    print(f"\nDone. Updated: {updated or '(none)'}  Skipped: {skipped or '(none)'}")
    return 0


if __name__ == "__main__":
    sys.exit(main())

