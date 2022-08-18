## K8s API aggregation guide

1. Aggregated API server는 HTTPS server 이어야 함
2. Aggregated API server의 endpoints는 아래와 같은 형식으로 구성 해야함
- https://(k8s-apiserver)/apis/{custom-group}/{custom-version} 에 대한 handlefunc 작성 (response로 API resource 정보)
- https://(k8s-apiserver)/apis/{custom-group}/{custom-version}/... 로 endpoint 구성
3. K8s Apiservice 객체 생성
- Aggregated API server pod의 svc 정보 추가
- CaBundle 설정 

