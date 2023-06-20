# CCA Vijayapurax

## Prerequisite
```
node & yarn,
go,
mongo server,
redis server,
```

## install modules
`yarn`

## run dev server
`yarn start`


### GCP storage setup
## --- steps for attaching gcp cloud bucker to vm ---

# -- run below command as same user in which application runs (for now root) --
```
gcloud auth application-default login
gcloud auth login
```

# -- add below in /etc/fstab ---
```
cca-public-india    /home/sathyanitsme/cdn gcsfuse rw,allow_other,x-systemd.requires=network-online.target,uid=0,gid=1002,file_mode=0777,dir_mode=0777
cca-private-india /home/sathyanitsme/storage gcsfuse rw,allow_other,x-systemd.requires=network-online.target,uid=0,gid=1002,file_mode=0777,dir_mode=0777
```

# -- add below to crontab (command crontab -e) to make sure the drive mounted --
```
@reboot sleep 60 && sudo systemctl daemon-reload && sudo systemctl restart local-fs.target
```

# -- run below command to restart file service --

```
umount /home/sathyanitsme/cdn
umount /home/sathyanitsme/storage
systemctl daemon-reload
systemctl restart local-fs.target
```

