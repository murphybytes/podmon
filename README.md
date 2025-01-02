# Podmon
### A little package that keeps track of the coming and going of Kubernetes pods.

[![Built with Devbox](https://www.jetify.com/img/devbox/shield_galaxy.svg)](https://www.jetify.com/devbox/docs/contributor-quickstart/)

I made this because I needed to be able to address a set of pods directory by the individual IP address of each pod. 

## Usage 
The examples directory has a sample application that shows how to use the package. You can run the example if you have devbox installed as follows:
```
devbox run minikube
```
Once minikube is running, start devbox shell and apply the `deploy.yaml` to install some busybox pods.
```
kubectl apply -f deploy.yaml -n foo
```
You can run the example application from the devbox shell and see pods come and go as we scale the deployment up and down. 


