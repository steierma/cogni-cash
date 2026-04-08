#!/usr/bin/env python3
"""
organize_payslip_history.py
----------------------------
Reads payslips_import.json and moves every referenced PDF that exists in
the history/ directory into a year-based subdirectory:

  history/Entgeltnachweis_2022_11_30.pdf  →  history/2022/Entgeltnachweis_2022_11_30.pdf

Usage (from repo root):
  python3 backend/scripts/organize_payslip_history.py

Or with explicit paths:
  python3 backend/scripts/organize_payslip_history.py \
      --json  backend/payslips/history/payslips_import.json \
      --history backend/payslips/history
"""

import argparse
import json
import shutil
import sys
from pathlib import Path


def main() -> None:
    repo_root = Path(__file__).resolve().parents[2]

    parser = argparse.ArgumentParser(description="Organise payslip history PDFs into year subdirectories.")
    parser.add_argument(
        "--json",
        default=str(repo_root / "backend" / "payslips" / "history" / "payslips_import.json"),
        help="Path to payslips_import.json",
    )
    parser.add_argument(
        "--history",
        default=str(repo_root / "backend" / "payslips" / "history"),
        help="Path to the history/ directory containing the PDFs",
    )
    parser.add_argument(
        "--dry-run",
        action="store_true",
        help="Print what would happen without moving any files",
    )
    args = parser.parse_args()

    json_path = Path(args.json)
    history_dir = Path(args.history)

    if not json_path.exists():
        print(f"ERROR: JSON file not found: {json_path}", file=sys.stderr)
        sys.exit(1)

    if not history_dir.is_dir():
        print(f"ERROR: History directory not found: {history_dir}", file=sys.stderr)
        sys.exit(1)

    with open(json_path, encoding="utf-8") as f:
        entries = json.load(f)

    moved = 0
    skipped_missing = 0
    skipped_already_organised = 0
    errors = 0

    for entry in entries:
        filename = entry.get("original_file_name", "")
        year = entry.get("period_year")

        if not filename or not year:
            print(f"  WARN  skipping entry with missing filename or year: {entry}")
            continue

        src = history_dir / filename
        dest_dir = history_dir / str(year)
        dest = dest_dir / filename

        if not src.exists():
            # Already moved, or never existed here
            if dest.exists():
                skipped_already_organised += 1
            else:
                print(f"  MISS  {filename}  (not in history/ or {year}/)")
                skipped_missing += 1
            continue

        if args.dry_run:
            print(f"  DRY   {filename}  →  history/{year}/{filename}")
            moved += 1
            continue

        try:
            dest_dir.mkdir(parents=True, exist_ok=True)
            shutil.move(str(src), str(dest))
            print(f"  MOVE  {filename}  →  history/{year}/{filename}")
            moved += 1
        except Exception as exc:
            print(f"  ERR   {filename}: {exc}", file=sys.stderr)
            errors += 1

    label = "Would move" if args.dry_run else "Moved"
    print(
        f"\n{label}: {moved}  |  Already organised: {skipped_already_organised}"
        f"  |  Missing: {skipped_missing}  |  Errors: {errors}"
    )

    if errors:
        sys.exit(1)


if __name__ == "__main__":
    main()

