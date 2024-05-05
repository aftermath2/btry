FROM lightninglabs/lnd:v0.17.4-beta

ENV LOCAL_USER_ID=1000

ENV LOCAL_GROUP_ID=1000

RUN adduser --disabled-password lnd

# Set the expected local user ID for shared group to access tor cookie
RUN apk --no-cache --no-progress add shadow=~4 sudo=~1 && \
    usermod -u $LOCAL_USER_ID lnd && \
    groupmod -g $LOCAL_GROUP_ID lnd

COPY entrypoint.sh /root/entrypoint.sh

COPY lnd.conf /tmp/lnd.conf

ENTRYPOINT [ "/root/entrypoint.sh" ]
