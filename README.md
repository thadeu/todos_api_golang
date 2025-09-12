# Organize Golang

Criar um alias de bin's para o local/bin do usuário, pra facilitar rodar os testes

```bash
sudo ln -fs $(pwd)/bin/gotest /usr/local/bin/gotest

sudo ln -fs $(pwd)/bin/gocover /usr/local/bin/gocover
```

## Branchs

https://medium.com/@smart_byte_labs/organize-like-a-pro-a-simple-guide-to-go-project-folder-structures-e85e9c1769c2

```sh
main: Tem último ajuste feito.

002_layered_structure: App feita em layers, a mais comum entre aplicações Go.

003_domain_driven_design: App feita em DDD, se quiser que fique complexo mas com código desacoplado e modular.

004_clean_arch: É quase como layers, mas com dependencias explicitas

005_module_structure: Cada pasta tem sua propria estrutura, handlers, services, models e logicas

006_feature_based: Mistura de Layers e Modular, o tem o internal e o pkg, mas o internal contém uma modular dentro dela

007_hexagonal: Famoso Ports e Adapters, dificil explicar, mas é uma mistura de tudo

008_monorepo: Pasta services, contém projetos individuais, com cmd, a layers ou modular

009_cqrs: Internal tem comands, queries, models, repositories e services

010_union: Internal possui domain, application(usecase), infra (persistence, api, websocket, etc..)
```