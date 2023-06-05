sudo mount -t gcsfuse -o rw,allow_other ccast "$HOME/cdn"
sudo mount -t gcsfuse -o rw,allow_other cca-private "$HOME/storage"
npm run build
./go_server