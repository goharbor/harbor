FROM cicd.harbor.bitsf.xin/harbor-depend/nginx:base

EXPOSE 8080

STOPSIGNAL SIGQUIT

HEALTHCHECK CMD curl --fail -s http://127.0.0.1:8080 || exit 1

USER nginx

CMD ["nginx", "-g", "daemon off;"]
