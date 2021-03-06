---
title: Использование minikube
sidebar: doc_sidebar
permalink: minikube_for_kube.html
folder: kube
---

Чтобы использовать dapp для деплоя образов в minikube:

* Собрать требуемые образы на хост-машине.
* Поднять minikube с docker-registry и proxy на хост-машине, см. [`dapp kube minikube setup`](#dapp-kube-minikube-setup).
* Загрузить собранные образы в docker-registry, указывая `:minikube` в качестве параметра `REPO`, через [`dapp dimg push :minikube`](base_commands.html#dapp-dimg-push).
* Применить конфигурацию kubernetes, указывая `:minikube` в качестве параметра `REPO`, через [`dapp kube deploy :minikube`](deploy_for_kube.html#dapp-kube-deploy).

### dapp kube minikube setup

```
dapp kube minikube setup
```

* Запускает minikube, принудительно перезапускает, если уже был запущен.
* Дожидается готовности кластера kubernetes в minikube.
* Запускает docker registry в minikube.
* Запускает в системе proxy для docker-registry по адресу `localhost:5000`.
  * Proxy пробрасывает прямо в pod docker-registry внутри minikube.
    * Как следствие при падении pod-а команду setup надо запускать заново.

