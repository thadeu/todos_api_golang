# Organize Golang

Create an alias from bin's to /usr/loca/bin to easy run tests via CLI

```bash
sudo ln -fs $(pwd)/bin/gotest /usr/local/bin/gotest

sudo ln -fs $(pwd)/bin/gocover /usr/local/bin/gocover
```

## Branchs

https://medium.com/@smart_byte_labs/organize-like-a-pro-a-simple-guide-to-go-project-folder-structures-e85e9c1769c2

```sh
main: Latest changes

002_layered_structure: Build with layers, most common Golang applications

003_domain_driven_design: DDD Struct

004_clean_arch: Clean Archicture basic application

007_hexagonal: Ports e Adapters structure
```