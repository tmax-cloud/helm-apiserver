# helm-apiserver

> Helm API server for HyperCloud Service

## Install helm-apiserver
1. Helm-apiserver를 설치하기 위한 네임스페이스를 생성
- kubectl create namespace helm-ns
2. Helm repository 저장을 위한 PVC 생성
- kubectl apply -f pvc.yaml ([파일](./deploy/pvc.yaml))
3. Helm-apiserver 생성
- kubectl apply -f deploy.yaml ([파일](./deploy/deploy.yaml))

## Helm-apiserver API 요약

| 리소스 | POST | GET | PUT | DELETE |
|:------- |:-------|:------- |:-------|:-------|
| /repos/{repo-name} | O | O | O | O |
| /charts/{chart-name| X | O | X | X |
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
- POST /releases
  * 요청된 Helm chart install
- GET /releases
  * 설치된 Helm release list 반환
- GET /releases/{release-name}
  * {release-name}의 Helm release 정보 반환
- PUT /releases/{release-name}
  * {release-name}의 Helm release update
- DELETE /releases/{release-name}
  * {release-name}의 Helm release uninstall

