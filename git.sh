git add .
if [ "$1" = "-m" ]; then
    message="$2"
fi
git commit -m "$message"
git push -f origin main
git tag v1.83
git push origin v1.83
