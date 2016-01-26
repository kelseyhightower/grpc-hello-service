# hello service demo

## Cleanup

```
$ kubectl delete rc hello-server auth-server
$ kubectl delete svc auth-server hello-server
$ kubectl delete secrets auth-server-tls hello-server-tls
```
```
$ gcloud compute disks delete auth-data
```

## Create the Auth Data Volume

```
$ gcloud compute disks create auth-data
```

## Create Secrets

```
kubectl get secrets
```

### Auth Server

```
conf2kube -n auth-server-tls -f auth-server-key.pem -k key.pem | kubectl create -f -
```
```
kubectl patch secret auth-server-tls -p `conf2kube -n auth-server-tls -f auth-server.pem -k cert.pem`
kubectl patch secret auth-server-tls -p `conf2kube -n auth-server-tls -f ca.pem -k ca.pem`
```
```
kubectl describe secrets auth-server-tls
```

### Hello Server

```
conf2kube -n hello-server-tls -f hello-server-key.pem -k key.pem | kubectl create -f -
```

```
kubectl patch secret hello-server-tls -p `conf2kube -n hello-server-tls -f hello-server.pem -k cert.pem`
kubectl patch secret hello-server-tls -p `conf2kube -n hello-server-tls -f ca.pem -k ca.pem`
kubectl patch secret hello-server-tls -p `conf2kube -n hello-server-tls -f auth-server.pem -k jwt.pem`
```
```
kubectl describe secrets hello-server-tls
```

## Create Replication Controllers

```
kubectl create -f auth-controller.yaml
```

```
kubectl get pods --watch
kubectl logs auth-server-xxxx
```

```
kubectl exec -i -t -p auth-server-xxxxx -c auth-server /bin/ash
```

```
/auth-admin -a -e kelsey.hightower@gmail.com -u kelseyhightower
exit
```

## Get auth token

```
auth-client
```

```
kubectl port-forward auth-server-xxxxx 7801:7801 7800:7800
```


## Deploy Hello server

```
kubectl create -f hello-controller.yaml
```

### Say Hello

```
kubectl port-forward hello-server-xxxxx 7901:7901 7900:7900
```

```
hello-client
```

## Create Services

```
kubectl create -f hello-service.yaml
kubectl create -f auth-service.yaml
```

