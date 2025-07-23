git add .
if [ "$1" = "-m" ]; then
    message="$2"
fi
git commit -m "$message"
git push -f origin main
git tag v0.1.11
git push origin v0.1.12
