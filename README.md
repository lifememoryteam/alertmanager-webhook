# alertmanager-webhook
Receive alertmanager webhook and create github issue


## How to Use
```
docker build -t alertmanager-webhook .
export GITHUB_TOKEN=xxxxx
docker run --name alertmanager-webhook01 -it -d -p 8000:8000 -env GITHUB_TOKEN=$GITHUB_TOKEN alertmanager-webhook
```

## Settings
Please set enviroment variable GITHUB_TOKEN and change repository owner and name.
```webhook.yaml
---
github:
  token: $GITHUB_TOKEN
  repository:
    owner: "ak1ra24"
    name: "samplerepo"
```