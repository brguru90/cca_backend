sudo -E bash -c "gcsfuse  ccast /home/sathyanitsme/cdn"
sudo -E bash -c "gcsfuse  cca-private /home/sathyanitsme/storage"
./go_server

sudo mount -t gcsfuse -o rw,allow_other  ccast "/home/sathyanitsme/cdn"