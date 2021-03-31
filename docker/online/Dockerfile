FROM 10.100.20.111:5000/official/golang:v1.11.2
MAINTAINER luwei <luwei@zhangyue.com>

ADD . /root/go/src/jiange

RUN mkdir -p /data/www/jiange /data/logs/uc.jiange.http /root/go/src/jiange \
    && cd /root/go/src/jiange \
    && source /etc/bashrc \
    && /usr/local/go/bin/go build -v -o /data/www/jiange/jiange main.go \
    && mv ./config_*.toml /data/www/jiange/ \
    && cp ./docker/online/supervisord.conf /data/www/jiange/

WORKDIR /data/www/jiange

CMD ["supervisord", "-c", "/data/www/jiange/supervisord.conf"]
