git add .
if [ "$1" = "-m" ]; then
    message="$2"
fi
if [ "$3" = "-v" ]; then
    ver="$4"
fi
git commit -m "$message"
# git push -f origin main
# git tag v"$ver"
# git push origin v"$ver"
