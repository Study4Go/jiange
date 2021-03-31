TAG=$(shell git tag | sort -bt"." -k1,1 -k2,2n -k3,3n |tail -n 1)

DEV_DOCKER_REGISTRIES=192.168.6.92:5000
GRAY_DOCKER_REGISTRIES=192.168.7.120:5000

DOCKER=docker
GO_JIANGE_IMAGE=${GRAY_DOCKER_REGISTRIES}/uc_base/jiange:${TAG}

gray:
	${DOCKER} build -t ${GO_JIANGE_IMAGE} .
	${DOCKER} rm -f www_jiange_1
	${DOCKER} run --name=www_jiange_1 -d \
        --ipc=host --net=host \
        -v /etc/localtime:/etc/localtime:ro \
        -v /data/www/jiange:/go/src/jiange \
        -v /data/logs/uc.jiange.http:/data/logs/uc.jiange.http \
        -v /tmp:/tmp \
        -e SUPERVISORD_LOG=/tmp/jiange_20000.log \
        -e SUPERVISORD_SOCK=/tmp/jiange_20000.sock \
        -e ZYAGENT_HTTPPORT=20000 \
        -e RUNMODE=prod -e NUM_PROCS_START=20000 \
	${GO_JIANGE_IMAGE}
