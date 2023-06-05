sudo mount -t gcsfuse -o rw,allow_other ccast "/home/sathyanitsme/cdn"
sudo mount -t gcsfuse -o rw,allow_other cca-private "/home/sathyanitsme/storage"
npm run build
./go_server