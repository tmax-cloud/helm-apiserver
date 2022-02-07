## Helm-apiserver API

| 리소스 | POST | GET | PUT | DELETE |
|:------- |:-------|:------- |:-------|:-------|
| /repos/{repo-name} | O | O | O | O |
| /charts/{chart-name}| X | O | X | X |
| /releases/{release-name} | O | O | O | O |