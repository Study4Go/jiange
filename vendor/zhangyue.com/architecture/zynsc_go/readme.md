### 简介
---
zynsc 是zhangyue name service client的缩写

### go client使用说明
---

#### [项目地址](http://192.168.6.70/architecture/zynsc_go)

#### 安装方法

1. 配置/etc/hosts

```sh
   # 开发环境
   echo "192.168.6.70 zhangyue.com" >> /etc/hosts
   # 生产环境
   echo "192.168.7.72 zhangyue.com" >> /etc/hosts
```

2. 安装和配置glide

```sh
# 安装glide
go get github.com/Masterminds/glide
```

glide的配置举例如下：

```
package: .
import:
- package: zhangyue.com/architecture/zynsc_go
  vcs: git
  repo: http://zhangyue.com/architecture/zynsc_go.git
  subpackages:
    - utils
    - qconf
```

3. 配置Makefile

```
GOPATH ?= $(shell go env GOPATH)
GO_BIN := $(shell which go)

ifeq "$(GOPATH)" ""
	$(error Please set the environment variable GOPATH before running `make`)
endif

CURDIR := $(shell pwd)
path_to_add := $(addsuffix /bin,$(subst :,/bin:,$(CURDIR)/_vendor:$(GOPATH)))

GO := go
GOBUILD := GOPATH=$(CURDIR)/_vendor:$(GOPATH) CGO_ENABLED=1 $(GO) build $(BUILD_FLAG)

export PATH := $(path_to_add):$(PATH)

build:
	$(GOBUILD)

update:
	which glide >/dev/null || curl https://glide.sh/get | sh
	which glide-vc || go get -v -u github.com/sgotti/glide-vc
	rm -r vendor && mv _vendor/src vendor || true
	rm -rf _vendor
ifdef PKG
	glide get -s -v --skip-test ${PKG}
else
	glide update -s -v -u --skip-test
endif
	@echo "removing test files"
	glide vc --only-code --no-tests
	mkdir -p _vendor
	mv vendor _vendor/src
```

4. 包的导入使用举例：

```
import "zhangyue.com/architecture/zynsc_go"
```

5. 安装和构建

```
make update
make build
```

#### 项目配置说明

* ZKAPI_PREFIX：用来配置zookeeper前缀
* ZKAPI_PATH: zkapi的服务提供者路径
* ZKAPI_ADD_CUSUMER: 服务消费者接口路径

#### 接口使用

* 接口名称: GetService
* 接口描述：获取服务提供者逻辑
* 接口参数说明
	* namespace: 名字空间,{group}.{service}.{type}
	* namespace: str
	* path: 路径
	* path: str
	* algorithm: 算法选择,参数非必选,默认为加权随机
	* algorithm: str
* 接口返回值说明
	* 返回Service
* GetService接口使用举例

```go
nscClient := zynsc.NewNSC("127.0.0.1", "8080")
// wr 表示加权随机的意思
str := nscClient.GetService("webyf.zywap2.http", "/" , "wr")
```

#### 算法模块说明

* Random (r)
随机选择一个服务提供者
* WeightRandom (wr)
加权随机，会根据服务提供者的权重进行加权随机
* SourceHashing (sr)
基于源站的哈希算法来选择服务提供者,每个服务器获取的服务提供者都是固定的

如果没在r,wr,sr这三个配置中则使用random算法
