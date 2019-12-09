operator-sdk generate k8s
operator-sdk generate openapi
kubectl apply -f deploy/crds/app.takowasa.net_clusterjobs_crd.yaml
kubectl apply -f deploy/crds/app.takowasa.net_v1alpha1_clusterjob_cr.yaml
operator-sdk up local --namespace=default