git add .
if [ "$1" = "-m" ]; then
    message="$2"
fi
git commit -m "$message"
git push -f origin main
git tag v0.1.13
git push origin v0.1.13
