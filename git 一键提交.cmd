git add .
git commit -m "auto commit for version code "
git push origin main


git remote rm github
git remote add github https://github.com/neko233-com/config233-go.git
git push github main