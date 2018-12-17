FROM photon:2.0

RUN tdnf install -y nginx-1.15.7-1.el7_4.ngx.x86_64 >> /dev/null\
    && ln -sf /dev/stdout /var/log/nginx/access.log \
    && ln -sf /dev/stderr /var/log/nginx/error.log \
    && tdnf clean all

#COPY nginx.repo /etc/yum.repos.d/
#RUN yum install -y nginx-1.15.7-1.el7_4.ngx.x86_64

EXPOSE 80
VOLUME /var/cache/nginx /var/log/nginx /run
STOPSIGNAL SIGQUIT

HEALTHCHECK CMD curl --fail -s http://127.0.0.1 || exit 1

CMD ["nginx", "-g", "daemon off;"]
