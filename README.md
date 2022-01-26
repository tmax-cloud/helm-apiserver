# helm-apiserver

> Helm API server for HyperCloud Service

## Install helm-apiserver
1. Helm-apiserver를 설치하기 위한 네임스페이스를 생성
```shell
kubectl create namespace helm-ns
```
2. Helm repository 저장을 위한 PVC 생성
```shell
kubectl apply -f pvc.yaml ([파일](./deploy/pvc.yaml))
```
3. Helm-apiserver 생성
```shell
kubectl apply -f deploy.yaml ([파일](./deploy/deploy.yaml))
```
