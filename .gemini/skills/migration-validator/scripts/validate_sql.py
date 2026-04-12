import os
import sys
import re

def check_migration(file_path):
    with open(file_path, 'r') as f:
        content = f.read().upper()
    
    issues = []
    
    # Check for CREATE TABLE without IF NOT EXISTS
    if "CREATE TABLE" in content and "IF NOT EXISTS" not in content:
        issues.append("CREATE TABLE missing IF NOT EXISTS")
        
    # Check for CREATE INDEX without IF NOT EXISTS
    if "CREATE INDEX" in content and "IF NOT EXISTS" not in content:
        issues.append("CREATE INDEX missing IF NOT EXISTS")
        
    # Check for INSERT without ON CONFLICT
    if "INSERT INTO" in content and "ON CONFLICT" not in content:
        issues.append("INSERT INTO missing ON CONFLICT (likely not idempotent)")
        
    # Check for SERIAL (should use UUID)
    if "SERIAL" in content:
        issues.append("Use of SERIAL detected. Use UUID PRIMARY KEY DEFAULT gen_random_uuid() instead.")
        
    # Check for DOWN migrations (CogniCash doesn't allow them in the same file)
    if "-- +GOOSE DOWN" in content or "DROP TABLE" in content:
        issues.append("Potential DOWN migration or DROP TABLE detected. CogniCash uses single-action Up-only migrations.")

    return issues

def main():
    migration_dir = "backend/migrations/"
    if not os.path.exists(migration_dir):
        print(f"Error: {migration_dir} not found.")
        sys.exit(1)
        
    all_issues = {}
    for filename in sorted(os.listdir(migration_dir)):
        if filename.endswith(".sql"):
            path = os.path.join(migration_dir, filename)
            issues = check_migration(path)
            if issues:
                all_issues[filename] = issues
                
    if all_issues:
        print("Migration Validation Failed:")
        for file, issues in all_issues.items():
            print(f"\n[{file}]")
            for issue in issues:
                print(f"  - {issue}")
        sys.exit(1)
    else:
        print("Success: All migrations are idempotent and follow standards.")

if __name__ == "__main__":
    main()
