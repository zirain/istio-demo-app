## # download kubebuilder and install locally

```
curl -L -o kubebuilder https://go.kubebuilder.io/dl/latest/$(go env GOOS)/$(go env GOARCH)
chmod +x kubebuilder && mv kubebuilder /usr/local/bin/
```

```
cd $(go env GOPATH)/src
mkdir -p zirain.dev/istio-demo-app
cd zirain.dev/istio-demo-app
go mod init zirain.dev/istio-demo-app
kubebuilder init --domain zirain.dev --repo zirain.dev/istio-demo-app
```


