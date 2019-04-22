#!/bin/sh

DEPLOY_FILE=deploy/proxier.yaml

rm $DEPLOY_FILE 2>/dev/null
touch $DEPLOY_FILE
for file in ./deploy/cluster_role.yaml ./deploy/cluster_role_binding.yaml ./deploy/crds/maegus_v1beta1_proxier_crd.yaml ./deploy/operator.yaml ./deploy/service_account.yaml
do
    echo "---" >> $DEPLOY_FILE
    cat $file >> $DEPLOY_FILE
done
