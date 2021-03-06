# helm-apiserver

> Helm API server for HyperCloud Service

## 개요
HyperCloud에서 Kubernetes package manager인 Helm service를 쉽게 사용할 수 있도록 지원하는 API server.
Helm Chart가 포함된 Helm Repository를 등록 후 차트 조회가 가능하며, 해당 차트를 사용하여 릴리즈를 생성함으로써 차트를 설치한다.

## Install helm-apiserver
1. Helm-apiserver를 설치하기 위한 네임스페이스를 생성
- kubectl create namespace helm-ns
2. Helm repository 저장을 위한 PVC 생성
- kubectl apply -f pvc.yaml ([파일](./deploy/pvc.yaml))
3. Helm-apiserver 생성
- kubectl apply -f deploy.yaml ([파일](./deploy/deploy.yaml))
4. HyperCloud API gateway사용을 위한 ingress 생성
- kubectl apply -f ingress.yaml ([파일](./deploy/ingress.yaml))

## Helm-apiserver API 요약
**자세한 내용은 [API docs 참조](https://documenter.getpostman.com/view/16732594/UVeGsSUr)**

| 리소스 | POST | GET | PUT | DELETE |
|:------- |:-------|:------- |:-------|:-------|
| /repos/{repo-name} | O | O | O | O |
| /charts/{chart-name}| X | O | X | X |
| /releases/{release-name} | O | O | O | O |

**/repos**
- POST /repos
  * Helm Repository 추가
- GET /repos
  * 추가된 Helm Repository list 반환
- PUT /repos
  * 추가된 Helm Repository의 chart sync
- DELETE /repos/{repo-name}
  * {repo-name} Repository 삭제

**/charts**
- POST, PUT, DELETE
  * 해당 Helm Repository에서 Chart 추가, 업데이트, 삭제
- GET /charts
  * 설치 가능한 chart list 반환
- GET /charts/{chart-name}
  * {chart-name}의 chart index.yaml 및 values.yaml 정보 반환

**/releases**
- Namespace scope 리소스이므로 Namespace path 있음
- POST /ns/{ns-name}/releases
  * 요청된 Helm chart install
- GET /ns/{ns-name}/releases
  * 설치된 Helm release list 반환
- GET /ns/{ns-name}/releases/{release-name}
  * {release-name}의 Helm release 정보 반환
- PUT /ns/{ns-name}/releases/{release-name}
  * {release-name}의 Helm release update
- DELETE /ns/{ns-name}/releases/{release-name}
  * {release-name}의 Helm release uninstall

