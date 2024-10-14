FROM node:22-alpine AS build
WORKDIR /app

COPY ./ui/package*.json ./

# Install dependencies
RUN npm install --frozen-lockfile

COPY ./ui ./
RUN npm run build

FROM nginx:alpine
COPY --from=build /app/build /usr/share/nginx/html

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
