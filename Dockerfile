FROM debian:11 as make-guardian-image

COPY bin/guardian /app/bin/

ENTRYPOINT [ "/app/bin/guardian" ]