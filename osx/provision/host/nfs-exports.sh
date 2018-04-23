#!/bin/bash

uid="$(id | grep -oP "uid=[0-9]+" | cut -d'=' -f2)"
gid="$(id | grep -oP "gid=[0-9]+" | cut -d'=' -f2)"

hostExports=(
  "\"$HOME/.dapp/builds\" 172.28.128.3(rw,no_subtree_check,all_squash,anonuid=$uid,anongid=$gid)"
  "\"/tmp\" 172.28.128.3(rw,no_subtree_check,all_squash,anonuid=$uid,anongid=$gid)"
  "\"$GOPATH\" 172.28.128.3(rw,no_subtree_check,all_squash,anonuid=$uid,anongid=$gid)"
)

for hostExportLine in "${hostExports[@]}" ; do
  grep -q "^$hostExportLine$" /etc/exports || echo $hostExportLine | sudo tee -a /etc/exports
done
