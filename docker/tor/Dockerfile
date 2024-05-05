FROM alpine:latest

ENV LOCAL_USER_ID=1000

RUN apk --no-cache --no-progress add curl tor shadow=~4 sudo=~1

COPY entrypoint.sh /root/entrypoint.sh

COPY torrc /tmp/torrc

HEALTHCHECK --interval=60s --timeout=15s --start-period=20s \
	CMD curl -x socks5h://127.0.0.1:9050 'https://check.torproject.org/api/ip' | grep -qm1 -E '"IsTor"\s*:\s*true'

ENTRYPOINT [ "/root/entrypoint.sh" ]
