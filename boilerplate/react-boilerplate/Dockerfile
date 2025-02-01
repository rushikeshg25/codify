FROM node:20 AS base

FROM base AS final
WORKDIR /app

WORKDIR /app/codeground
COPY codeground/package*.json ./
RUN npm install
COPY codeground ./

WORKDIR /app/wss
COPY wss/package*.json ./
RUN npm install
COPY wss ./

EXPOSE 5173 9000

WORKDIR /app
RUN echo '#!/bin/sh\n\
cd /app/wss && node index.js & \n\
cd /app/codeground && \n\
# Keep container running\n\
tail -f /dev/null' > /app/start.sh && \
    chmod +x /app/start.sh

CMD ["sh", "-c", "/app/start.sh"]