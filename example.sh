#!/usr/bin/env bash
set -euo pipefail

# Build the binary
echo "=== Building prd ==="
go build -o prd_bin .

PRD=$(pwd)/prd_bin

# Create a fresh example directory and work inside it
rm -rf example
mkdir example
cd example

# Initialize a git repo (needed for feat start/current)
git init -q
git commit --allow-empty -m "init" -q

echo ""
echo "=== feat new: create two features ==="
$PRD feat new --name "Authentication" --description "User login and registration"
$PRD feat new --name "Dashboard" --description "Main dashboard view"

echo ""
echo "=== feat ls: list features ==="
$PRD feat ls

echo ""
echo "=== us new: create user stories for feature 001 ==="
$PRD feat 001 us new --name "Login form" --description "User can log in" \
  --acceptance-criteria "Email field present" --acceptance-criteria "Password field present" \
  --technical-considerations "Use bcrypt for passwords"
$PRD feat 001 us new --name "Registration form" --description "User can register" \
  --acceptance-criteria "Name field" --acceptance-criteria "Email field" --acceptance-criteria "Password field"
$PRD feat 001 us new --name "Password reset" --description "User can reset password" \
  --acceptance-criteria "Reset email sent" --acceptance-criteria "Token validated"

echo ""
echo "=== feat <id>: show feature PRD ==="
$PRD feat 001

echo ""
echo "=== us ls: list user stories ==="
$PRD feat 001 us ls

echo ""
echo "=== us <id>: show single user story ==="
$PRD feat 001 us 1

echo ""
echo "=== us next: get first incomplete story ==="
echo -n "Next story: "
$PRD feat 001 us next

echo ""
echo "=== accept ls: list acceptance criteria (read-only from prd.md) ==="
$PRD feat 001 us 1 accept ls

echo ""
echo "=== us complete: mark US-1 as done in progress.md ==="
$PRD feat 001 us 1 complete

echo ""
echo "=== us ls: verify US-1 is now [x] ==="
$PRD feat 001 us ls

echo ""
echo "=== us next: should now return US-2 ==="
echo -n "Next story: "
$PRD feat 001 us next

echo ""
echo "=== us complete: toggle US-1 back to incomplete ==="
$PRD feat 001 us 1 complete

echo ""
echo "=== us ls: verify US-1 is [ ] again ==="
$PRD feat 001 us ls

echo ""
echo "=== us complete: re-complete US-1 and complete US-2 ==="
$PRD feat 001 us 1 complete
$PRD feat 001 us 2 complete

echo ""
echo "=== feat ls: show updated progress counts ==="
$PRD feat ls

echo ""
echo "=== note add: add notes ==="
$PRD feat 001 note add --content "Decided to use JWT tokens for auth"
$PRD feat 001 note add --content "Need to add rate limiting to login endpoint"

echo ""
echo "=== note ls: list notes from progress.md ==="
$PRD feat 001 note ls

echo ""
echo "=== progress.md contents ==="
cat prd/001-authentication/progress.md

echo ""
echo "=== us delete: remove US-3 ==="
$PRD feat 001 us 3 delete

echo ""
echo "=== us ls: verify US-3 is gone ==="
$PRD feat 001 us ls

echo ""
echo "=== progress.md after delete ==="
cat prd/001-authentication/progress.md

echo ""
echo "=== bump: bump version ==="
$PRD feat 001 bump

echo ""
echo "=== feat start: create feature branch ==="
$PRD feat start 001

echo ""
echo "=== feat current: detect feature from branch ==="
echo -n "Current feature: "
$PRD feat current

echo ""
echo "=== skill: print skill doc (first 5 lines) ==="
$PRD skill | head -5

echo ""
echo "=== prd.md is unchanged by progress operations ==="
echo "--- prd.md ---"
cat prd/001-authentication/prd.md

echo ""
echo "=== All commands executed successfully ==="

# Cleanup
cd ..
# rm -rf example
# rm -f prd_bin
