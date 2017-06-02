# sync with latest master code.
git fetch upstream
git reset --hard upstream/drone
git config credential.helper store
git push origin drone-nightly

cd tests/nightly
tokenfile=nightly-tokenfile

# append string into the token file.
echo "$(date) $(ls -1 | wc -l)" >> $tokenfile
echo "------------------------" >> $tokenfile

cd ../..
git checkout drone-nightly
git add tests/nightly/$tokenfile
git commit -m "nightly run"
git config credential.helper store
git push origin drone-nightly --force
