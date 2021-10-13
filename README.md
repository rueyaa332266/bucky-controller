# Bucky-controller
Bucky-controller is a custom controller for the Bucky CRD on Kubernetes.

Bucky CRD helps you build the environment to execute Bucky based test scripts on your kubernetes cluster.

It will automatically create a deployment which contains one [Bucky-core]((https://github.com/lifull-dev/bucky-core)), one [Selenium-node](https://hub.docker.com/r/selenium/hub) and some multiple [Selenium-node-chrome](https://hub.docker.com/r/selenium/node-chrome).

## Bucky CRD sample
```yaml
apiVersion: buckycontroller.k8s.io/v1alpha1
kind: Bucky
metadata:
  name: bucky-sample
spec:
  seleniumNodeNumber: 3
  nodeInstanceNumber: 3
  buckyCoreImage: "bc:latest"
  buckyCommand: "bucky run -t e2e -d"
```

## Preparetion
We should make our own Bucky image which contains test scripts.

1. Move to your Bucky project

2. Create the Dockerfile
```Dockerfile
FROM lifullsetg/bucky-core:latest
COPY . /app
```

3. Adjust e2e config for Bucky controller
 - Change `selenium_ip` into `localhost`
 - Change `e2e_parallel_num` into `<%= ENV['E2E_PARALLEL_NUM'] %>`
```
# config/e2e_config.yml

:selenium_ip: localhost
:e2e_parallel_num: <%= ENV['E2E_PARALLEL_NUM'] %>
```

4. Build your Bucky image
```
docker build -t bc-sample:latest .
```


## Setup the Bucky-controller in your Cluster

> Note: Make sure you install kustomize first❗

1. Clone the source code.
```
git clone git@github.com:rueyaa332266/bucky-controller.git
cd bucky-controller
```

2. Install Bucky CRD into the cluster.
```
make install
```

3. Deploy Bucky-controller into the cluster.
```
export IMG=aa332266/bucky-controller:latest
make deploy
```

## Usage
After deploying a Bucky-controller in your cluster, you can use the Bucky CRD normally.

1. Create a Bucky CRD manifest. (You can just copy the Bucky CRD sample in README.)

2. Set the Selenium-node-chrome number, node instance number, Bucky image we prepared, and the Bucky command to run.
```yaml
apiVersion: buckycontroller.k8s.io/v1alpha1
kind: Bucky
metadata:
  name: bucky-sample
spec:
  seleniumNodeNumber: 3
  nodeInstanceNumber: 3
  buckyCoreImage: "bc-sample:latest"
  buckyCommand: "bucky run -t e2e -d"
```

3. Apply the manifest.
```
kubectl apply -f bucky.yml
```

Bucky-controller will create a pod and automatically execute the test scripts.

## Remove Bucky-controller
```
kustomize build config/default | kubectl apply -f -
kustomize build config/crd | kubectl delete -f -
```

## Build by
[Kubebuilder](https://github.com/kubernetes-sigs/kubebuilder)

