# Podmon
### A little package that keeps track of the coming and going of Kubernetes pods.

[![Built with Devbox](https://www.jetify.com/img/devbox/shield_galaxy.svg)](https://www.jetify.com/devbox/docs/contributor-quickstart/)

## Usage 
See the examples directory. In short we define a namespace and label for the pods 
we want to pay attention to. We get events on a channel if a matching pod is created or the 
ip address is modified, or if the pod is removed. 

