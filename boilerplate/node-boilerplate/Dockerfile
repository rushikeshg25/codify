FROM node:20 AS base

FROM base AS final

# The node image already ships an unprivileged uid 1000 ("node"). The pod
# security context pins the container to that uid, so /app has to be owned by
# it and every install has to run as it.
RUN mkdir -p /app/codeground /app/wss && chown -R node:node /app

USER node

WORKDIR /app/codeground
COPY --chown=node:node codeground/package*.json ./
RUN npm install
COPY --chown=node:node codeground ./

WORKDIR /app/wss
COPY --chown=node:node wss/package*.json ./
RUN npm install
COPY --chown=node:node wss ./

EXPOSE 8090 9000

WORKDIR /app
RUN echo '#!/bin/sh\n\
cd /app/wss && node index.js & \n\
cd /app/codeground && \n\
# Keep container running\n\
tail -f /dev/null' > /app/start.sh && \
    chmod +x /app/start.sh

CMD ["sh", "-c", "/app/start.sh"]
