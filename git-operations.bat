@echo off
cd /d "D:\study part\GoClawPet"
echo === Checking git status ===
git status --short
echo.
echo === Last 3 commits ===
git log --oneline -3
echo.
echo === Pushing to origin ===
git push origin dev
echo.
echo === Creating release tag v0.1.0 ===
git tag -f v0.1.0
git push origin v0.1.0
echo.
echo === Done ===
echo Please check: https://github.com/hazy-cloudy/GoClaw/actions
pause
