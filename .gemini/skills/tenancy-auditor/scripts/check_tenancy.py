import os
import sys
import subprocess

def check_tenancy():
    # Use ripgrep to find SQL queries in repositories
    # Look for SELECT/UPDATE/DELETE missing user_id
    search_dirs = ["backend/internal/adapter/repository/"]
    
    issues = []
    
    # Simple check for SELECT/UPDATE/DELETE without user_id in common repository files
    for root, dirs, files in os.walk(search_dirs[0]):
        for file in files:
            if file.endswith(".go") and not file.endswith("_test.go"):
                path = os.path.join(root, file)
                with open(path, 'r') as f:
                    content = f.read()
                    
                # Look for SQL strings
                # This is a naive regex but good for a basic check
                queries = re.findall(r'`([^`]+)`', content)
                for q in queries:
                    qu = q.upper()
                    if ("SELECT " in qu or "UPDATE " in qu or "DELETE " in qu) and "USER_ID" not in qu:
                        # Exclude common tables that might not have user_id (e.g., users table itself)
                        if "FROM USERS" not in qu and "UPDATE USERS" not in qu:
                            issues.append(f"{path}: Missing USER_ID in query: {q.strip()}")

    return issues

import re

if __name__ == "__main__":
    print("Checking tenancy isolation...")
    issues = check_tenancy()
    if issues:
        print("\nPotential Tenancy Issues Found:")
        for issue in issues:
            print(f"  - {issue}")
    else:
        print("Success: Basic tenancy check passed.")
