#!/bin/bash
HOST1=clab-evpn-leaf2
PWD=NokiaSrl1!
echo "update the interface description: system0"
curl https://admin:${PWD}@${HOST1}/jsonrpc -s --insecure -H "Content-Type:application/json" -d @diff_inf_update_system0.json | jq -r '.result[0]'

echo "Delete the interface description: mgmt0"
#curl https://admin:${PWD}@${HOST1}/jsonrpc -s --insecure -H "Content-Type:application/json" -d @diff_inf_delete_mgmt.json | jq -r '.result[0]' 
curl https://admin:${PWD}@${HOST1}/jsonrpc -s --insecure -H "Content-Type:application/json" -d @diff_inf_delete_mgmt.json | jq -r '.result[0]'

echo "Replace the interface description: ethernet-1/1.1"
curl https://admin:${PWD}@${HOST1}/jsonrpc -s --insecure -H "Content-Type:application/json" -d @diff_inf_replace_e1-1-1.json | jq -r '.result[0]' 
